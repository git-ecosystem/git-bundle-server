package bundles

import (
	"bufio"
	"encoding/json"
	"fmt"
	"git-bundle-server/internal/core"
	"os"
	"strconv"
	"strings"
	"time"
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

func addBundleToList(bundle Bundle, list BundleList) {
	list.Bundles[bundle.CreationToken] = bundle
}

func CreateInitialBundle(repo core.Repository) Bundle {
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

func CreateDistinctBundle(repo core.Repository, list BundleList) Bundle {
	timestamp := time.Now().UTC().Unix()

	_, c := list.Bundles[timestamp]

	for c {
		timestamp++
		_, c = list.Bundles[timestamp]
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

func SingletonList(bundle Bundle) BundleList {
	list := BundleList{1, "all", make(map[int64]Bundle)}

	addBundleToList(bundle, list)

	return list
}

// Given a BundleList
func WriteBundleList(list BundleList, repo core.Repository) error {
	listFile := repo.WebDir + "/bundle-list"
	jsonFile := repo.RepoDir + "/bundle-list.json"

	// TODO: Formalize lockfile concept.
	f, err := os.OpenFile(listFile+".lock", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failure to open file: %w", err)
	}

	out := bufio.NewWriter(f)

	fmt.Fprintf(
		out, "[bundle]\n\tversion = %d\n\tmode = %s\n\n",
		list.Version, list.Mode)

	for token, bundle := range list.Bundles {
		fmt.Fprintf(
			out, "[bundle \"%d\"]\n\turi = %s\n\tcreationToken = %d\n\n",
			token, bundle.URI, token)
	}

	out.Flush()
	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}

	f, err = os.OpenFile(jsonFile+".lock", os.O_WRONLY|os.O_CREATE, 0600)
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

func GetBundleList(repo core.Repository) (*BundleList, error) {
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

func GetBundleHeader(bundle Bundle) (*BundleHeader, error) {
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
			strings.HasSuffix(line, " git bundle\n") {
			header.Version, err = strconv.ParseInt(line[3:len(line)-len(" git bundle\n")], 10, 64)
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

func GetAllPrereqsForIncrementalBundle(list BundleList) ([]string, error) {
	prereqs := []string{}

	for _, bundle := range list.Bundles {
		header, err := GetBundleHeader(bundle)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bundle file %s: %w", bundle.Filename, err)
		}

		for _, oid := range header.Refs {
			prereqs = append(prereqs, "^"+oid)
		}
	}

	return prereqs, nil
}
