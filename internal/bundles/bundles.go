package bundles

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/github/git-bundle-server/internal/core"
	"github.com/github/git-bundle-server/internal/git"
	"github.com/github/git-bundle-server/internal/log"
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
	URI           string
	Filename      string
	CreationToken int64
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
	logger    log.TraceLogger
	gitHelper git.GitHelper
}

func NewBundleProvider(
	l log.TraceLogger,
	g git.GitHelper,
) BundleProvider {
	return &bundleProvider{
		logger:    l,
		gitHelper: g,
	}
}

func (b *bundleProvider) CreateInitialBundle(ctx context.Context, repo *core.Repository) Bundle {
	timestamp := time.Now().UTC().Unix()
	bundleName := "bundle-" + fmt.Sprint(timestamp) + ".bundle"
	bundleFile := repo.WebDir + "/" + bundleName
	bundle := Bundle{
		URI:           "./" + bundleName,
		Filename:      bundleFile,
		CreationToken: timestamp,
	}

	return bundle
}

func (b *bundleProvider) createDistinctBundle(repo *core.Repository, list *BundleList) Bundle {
	timestamp := time.Now().UTC().Unix()

	keys := list.sortedCreationTokens()

	maxTimestamp := keys[len(keys)-1]
	if timestamp <= maxTimestamp {
		timestamp = maxTimestamp + 1
	}

	bundleName := "bundle-" + fmt.Sprint(timestamp) + ".bundle"
	bundleFile := repo.WebDir + "/" + bundleName
	bundle := Bundle{
		URI:           "./" + bundleName,
		Filename:      bundleFile,
		CreationToken: timestamp,
	}

	return bundle
}

func (b *bundleProvider) CreateSingletonList(ctx context.Context, bundle Bundle) *BundleList {
	list := BundleList{1, "all", make(map[int64]Bundle)}

	list.addBundle(bundle)

	return &list
}

// Given a BundleList
func (b *bundleProvider) WriteBundleList(ctx context.Context, list *BundleList, repo *core.Repository) error {
	//lint:ignore SA4006 always override the ctx with the result from 'Region()'
	ctx, exitRegion := b.logger.Region(ctx, "bundles", "write_bundle_list")
	defer exitRegion()

	listFile := repo.WebDir + "/bundle-list"
	jsonFile := repo.RepoDir + "/bundle-list.json"

	// TODO: Formalize lockfile concept.
	f, err := os.OpenFile(listFile+".lock", os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("failure to open file: %w", err)
	}

	out := bufio.NewWriter(f)

	fmt.Fprintf(
		out, "[bundle]\n\tversion = %d\n\tmode = %s\n\n",
		list.Version, list.Mode)

	keys := list.sortedCreationTokens()

	for _, token := range keys {
		bundle := list.Bundles[token]
		fmt.Fprintf(
			out, "[bundle \"%d\"]\n\turi = %s\n\tcreationToken = %d\n\n",
			token, bundle.URI, token)
	}

	out.Flush()
	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	f, err = os.OpenFile(jsonFile+".lock", os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open JSON file: %w", err)
	}

	data, jsonErr := json.Marshal(list)
	if jsonErr != nil {
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

	f.Sync()
	f.Close()

	renameErr := os.Rename(jsonFile+".lock", jsonFile)
	if renameErr != nil {
		return fmt.Errorf("failed to rename JSON file: %w", renameErr)
	}

	return os.Rename(listFile+".lock", listFile)
}

func (b *bundleProvider) GetBundleList(ctx context.Context, repo *core.Repository) (*BundleList, error) {
	//lint:ignore SA4006 always override the ctx with the result from 'Region()'
	ctx, exitRegion := b.logger.Region(ctx, "bundles", "get_bundle_list")
	defer exitRegion()

	jsonFile := repo.RepoDir + "/bundle-list.json"

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

	bundle := Bundle{
		CreationToken: maxTimestamp,
		Filename:      fmt.Sprintf("%s/base-%d.bundle", repo.WebDir, maxTimestamp),
		URI:           fmt.Sprintf("./base-%d.bundle", maxTimestamp),
	}

	err := b.gitHelper.CreateBundleFromRefs(ctx, repo.RepoDir, bundle.Filename, refs)
	if err != nil {
		return err
	}

	list.Bundles[maxTimestamp] = bundle
	return nil
}
