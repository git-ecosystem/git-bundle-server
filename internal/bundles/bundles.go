package bundles

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/git"
	"github.com/github/git-bundle-server/internal/log"
)

const (
	BundleListJsonFilename string = "bundle-list.json"
	BundleListFilename     string = "bundle-list"
	RepoBundleListFilename string = "repo-bundle-list"
)

type BundleHeader struct {
	Version int64

	// The Refs map is given as Refs[<refname>] = <oid>.
	Refs map[string]string

	// The PrereqCommits map is given as
	// PrereqCommits[<oid>] = <commit-msg>
	PrereqCommits map[string]string
}

type Bundle struct {
	// The absolute path to the bundle from the root of the bundle web server,
	// typically '/org/route/filename'.
	URI string

	// The absolute path to the bundle on disk
	Filename string

	// The creation token used in Git's 'creationToken' heuristic
	CreationToken int64
}

func NewBundle(repo *core.Repository, timestamp int64) Bundle {
	bundleName := fmt.Sprintf("bundle-%d.bundle", timestamp)
	return Bundle{
		URI:           path.Join("/", repo.Route, bundleName),
		Filename:      filepath.Join(repo.WebDir, bundleName),
		CreationToken: timestamp,
	}
}

type BundleList struct {
	Version int
	Mode    string
	Bundles map[int64]Bundle
}

func (list *BundleList) addBundle(bundle Bundle) {
	list.Bundles[bundle.CreationToken] = bundle
}

func (list *BundleList) sortedCreationTokens() []int64 {
	keys := make([]int64, 0, len(list.Bundles))
	for timestamp := range list.Bundles {
		keys = append(keys, timestamp)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return keys
}

type BundleProvider interface {
	CreateInitialBundle(ctx context.Context, repo *core.Repository) Bundle
	CreateIncrementalBundle(ctx context.Context, repo *core.Repository, list *BundleList) (*Bundle, error)

	CreateSingletonList(ctx context.Context, bundle Bundle) *BundleList
	WriteBundleList(ctx context.Context, list *BundleList, repo *core.Repository) error
	GetBundleList(ctx context.Context, repo *core.Repository) (*BundleList, error)
	CollapseList(ctx context.Context, repo *core.Repository, list *BundleList) error
}

type bundleProvider struct {
	logger     log.TraceLogger
	fileSystem common.FileSystem
	gitHelper  git.GitHelper
}

func NewBundleProvider(
	l log.TraceLogger,
	fs common.FileSystem,
	g git.GitHelper,
) BundleProvider {
	return &bundleProvider{
		logger:     l,
		fileSystem: fs,
		gitHelper:  g,
	}
}

func (b *bundleProvider) CreateInitialBundle(ctx context.Context, repo *core.Repository) Bundle {
	return NewBundle(repo, time.Now().UTC().Unix())
}

func (b *bundleProvider) createDistinctBundle(repo *core.Repository, list *BundleList) Bundle {
	timestamp := time.Now().UTC().Unix()

	keys := list.sortedCreationTokens()

	maxTimestamp := keys[len(keys)-1]
	if timestamp <= maxTimestamp {
		timestamp = maxTimestamp + 1
	}

	return NewBundle(repo, timestamp)
}

func (b *bundleProvider) CreateSingletonList(ctx context.Context, bundle Bundle) *BundleList {
	list := BundleList{1, "all", make(map[int64]Bundle)}

	list.addBundle(bundle)

	return &list
}

// Given a BundleList, write the bundle list content to the web directory.
func (b *bundleProvider) WriteBundleList(ctx context.Context, list *BundleList, repo *core.Repository) error {
	//lint:ignore SA4006 always override the ctx with the result from 'Region()'
	ctx, exitRegion := b.logger.Region(ctx, "bundles", "write_bundle_list")
	defer exitRegion()

	var listLockFile, repoListLockFile, jsonLockFile common.LockFile
	rollbackAll := func() {
		if listLockFile != nil {
			listLockFile.Rollback()
		}
		if repoListLockFile != nil {
			repoListLockFile.Rollback()
		}
		if jsonLockFile != nil {
			jsonLockFile.Rollback()
		}
	}

	// Write the bundle list files: one for requests with a trailing slash
	// (where the relative bundle paths are '<bundlefile>'), one for requests
	// without a trailing slash (where the relative bundle paths are
	// '<repo>/<bundlefile>').
	keys := list.sortedCreationTokens()
	writeListFile := func(f io.Writer, requestUri string) error {
		out := bufio.NewWriter(f)
		defer out.Flush()

		fmt.Fprintf(
			out, "[bundle]\n\tversion = %d\n\tmode = %s\n\n",
			list.Version, list.Mode)

		uriBase := path.Dir(requestUri) + "/"
		for _, token := range keys {
			bundle := list.Bundles[token]

			// Get the URI relative to the bundle server root
			uri := strings.TrimPrefix(bundle.URI, uriBase)
			if uri == bundle.URI {
				panic("error resolving bundle URI paths")
			}

			fmt.Fprintf(
				out, "[bundle \"%d\"]\n\turi = %s\n\tcreationToken = %d\n\n",
				token, uri, token)
		}
		return nil
	}

	listLockFile, err := b.fileSystem.WriteLockFileFunc(
		filepath.Join(repo.WebDir, BundleListFilename),
		func(f io.Writer) error {
			return writeListFile(f, path.Join("/", repo.Route)+"/")
		},
	)
	if err != nil {
		rollbackAll()
		return err
	}

	repoListLockFile, err = b.fileSystem.WriteLockFileFunc(
		filepath.Join(repo.WebDir, RepoBundleListFilename),
		func(f io.Writer) error {
			return writeListFile(f, path.Join("/", repo.Route))
		},
	)
	if err != nil {
		rollbackAll()
		return err
	}

	// Write the (internal-use) JSON representation of the bundle list
	jsonLockFile, err = b.fileSystem.WriteLockFileFunc(
		filepath.Join(repo.RepoDir, BundleListJsonFilename),
		func(f io.Writer) error {
			data, err := json.Marshal(list)
			if err != nil {
				return fmt.Errorf("failed to convert list to JSON: %w", err)
			}

			written := 0
			for written < len(data) {
				n, writeErr := f.Write(data[written:])
				if writeErr != nil {
					return fmt.Errorf("failed to write JSON: %w", err)
				}
				written += n
			}

			return nil
		},
	)
	if err != nil {
		rollbackAll()
		return err
	}

	// Commit all lockfiles
	err = jsonLockFile.Commit()
	if err != nil {
		return fmt.Errorf("failed to rename JSON file: %w", err)
	}

	err = listLockFile.Commit()
	if err != nil {
		return fmt.Errorf("failed to rename bundle list file: %w", err)
	}

	err = repoListLockFile.Commit()
	if err != nil {
		return fmt.Errorf("failed to rename repo-level bundle list file: %w", err)
	}

	return nil
}

func (b *bundleProvider) GetBundleList(ctx context.Context, repo *core.Repository) (*BundleList, error) {
	//lint:ignore SA4006 always override the ctx with the result from 'Region()'
	ctx, exitRegion := b.logger.Region(ctx, "bundles", "get_bundle_list")
	defer exitRegion()

	jsonFile := filepath.Join(repo.RepoDir, BundleListJsonFilename)

	reader, err := os.Open(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	var list BundleList
	err = json.NewDecoder(reader).Decode(&list)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON from file: %w", err)
	}

	return &list, nil
}

func (b *bundleProvider) getBundleHeader(bundle Bundle) (*BundleHeader, error) {
	file, err := os.Open(bundle.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle file: %w", err)
	}

	header := BundleHeader{
		Version:       0,
		Refs:          make(map[string]string),
		PrereqCommits: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		buffer := scanner.Bytes()

		if len(buffer) == 0 ||
			buffer[0] == '\n' {
			break
		}

		line := string(buffer)

		if line[0] == '#' &&
			strings.HasPrefix(line, "# v") &&
			strings.HasSuffix(line, " git bundle") {
			header.Version, err = strconv.ParseInt(line[3:len(line)-len(" git bundle")], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bundle version: %s", err)
			}
			continue
		}

		if header.Version == 0 {
			return nil, fmt.Errorf("failed to parse bundle header: no version")
		}

		if line[0] == '@' {
			// This is a capability. Ignore for now.
			continue
		}

		if line[0] == '-' {
			// This is a prerequisite
			space := strings.Index(line, " ")
			if space < 0 {
				return nil, fmt.Errorf("failed to parse rerequisite '%s'", line)
			}

			oid := line[0:space]
			message := line[space+1 : len(line)-1]
			header.PrereqCommits[oid] = message
		} else {
			// This is a tip
			space := strings.Index(line, " ")

			if space < 0 {
				return nil, fmt.Errorf("failed to parse tip '%s'", line)
			}

			oid := line[0:space]
			ref := line[space+1 : len(line)-1]
			header.Refs[ref] = oid
		}
	}

	return &header, nil
}

func (b *bundleProvider) getAllPrereqsForIncrementalBundle(list *BundleList) ([]string, error) {
	prereqs := []string{}

	for _, bundle := range list.Bundles {
		header, err := b.getBundleHeader(bundle)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bundle file %s: %w", bundle.Filename, err)
		}

		for _, oid := range header.Refs {
			prereqs = append(prereqs, "^"+oid)
		}
	}

	return prereqs, nil
}

func (b *bundleProvider) CreateIncrementalBundle(ctx context.Context, repo *core.Repository, list *BundleList) (*Bundle, error) {
	ctx, exitRegion := b.logger.Region(ctx, "bundles", "create_incremental_bundle")
	defer exitRegion()

	bundle := b.createDistinctBundle(repo, list)

	lines, err := b.getAllPrereqsForIncrementalBundle(list)
	if err != nil {
		return nil, err
	}

	written, err := b.gitHelper.CreateIncrementalBundle(ctx, repo.RepoDir, bundle.Filename, lines)
	if err != nil {
		return nil, fmt.Errorf("failed to create incremental bundle: %w", err)
	}

	if !written {
		return nil, nil
	}

	return &bundle, nil
}

func (b *bundleProvider) CollapseList(ctx context.Context, repo *core.Repository, list *BundleList) error {
	ctx, exitRegion := b.logger.Region(ctx, "bundles", "collapse_list")
	defer exitRegion()

	maxBundles := 5

	if len(list.Bundles) <= maxBundles {
		return nil
	}

	keys := list.sortedCreationTokens()

	refs := make(map[string]string)

	maxTimestamp := int64(0)

	for i := range keys[0 : len(keys)-maxBundles+1] {
		bundle := list.Bundles[keys[i]]

		if bundle.CreationToken > maxTimestamp {
			maxTimestamp = bundle.CreationToken
		}

		header, err := b.getBundleHeader(bundle)
		if err != nil {
			return fmt.Errorf("failed to parse bundle file %s: %w", bundle.Filename, err)
		}

		// Ignore the old ref name and instead use the OID
		// to generate the ref name. This allows us to create new
		// refs that point to exactly these objects without disturbing
		// refs/heads/ which is tracking the remote refs.
		for _, oid := range header.Refs {
			refs["refs/base/"+oid] = oid
		}

		delete(list.Bundles, keys[i])
	}

	// TODO: Use Git to determine which OIDs are "maximal" in the set
	// and which are not implied by the previous ones.

	// TODO: Use Git to determine which OIDs are required as prerequisites
	// of the remaining bundles and latest ref tips, so we can "GC" the
	// branches that were never merged and may have been force-pushed or
	// deleted.

	bundle := NewBundle(repo, maxTimestamp)

	err := b.gitHelper.CreateBundleFromRefs(ctx, repo.RepoDir, bundle.Filename, refs)
	if err != nil {
		return err
	}

	list.Bundles[maxTimestamp] = bundle
	return nil
}
