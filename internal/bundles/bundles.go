package bundles

import (
	"bufio"
	"encoding/json"
	"fmt"
	"git-bundle-server/internal/core"
	"os"
	"time"
)

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
