package bundles_test

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/git-ecosystem/git-bundle-server/internal/bundles"
	"github.com/git-ecosystem/git-bundle-server/internal/core"
	. "github.com/git-ecosystem/git-bundle-server/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var writeBundleListTests = []struct {
	title string

	// Inputs
	bundleList *bundles.BundleList
	repo       *core.Repository

	// Expected values
	bundleListFile     []string
	repoBundleListFile []string

	// Expected output
	expectErr bool
}{
	{
		"Empty bundle list",
		&bundles.BundleList{
			Version:   1,
			Mode:      "all",
			Heuristic: "creationToken",
			Bundles:   map[int64]bundles.Bundle{},
		},
		&core.Repository{
			Route:   "test/repo",
			RepoDir: "/test/home/git-bundle-server/git/test/myrepo/",
			WebDir:  "/test/home/git-bundle-server/www/test/myrepo/",
		},
		[]string{
			`[bundle]`,
			`	version = 1`,
			`	mode = all`,
			`	heuristic = creationToken`,
			``,
		},
		[]string{
			`[bundle]`,
			`	version = 1`,
			`	mode = all`,
			`	heuristic = creationToken`,
			``,
		},
		false,
	},
	{
		"Single bundle list",
		&bundles.BundleList{
			Version:   1,
			Mode:      "all",
			Heuristic: "creationToken",
			Bundles: map[int64]bundles.Bundle{
				1: {
					URI:           "/test/myrepo/bundle-1.bundle",
					Filename:      "/test/home/git-bundle-server/www/test/myrepo/bundle-1.bundle",
					CreationToken: 1,
				},
			},
		},
		&core.Repository{
			Route:   "test/myrepo",
			RepoDir: "/test/home/git-bundle-server/git/test/myrepo/",
			WebDir:  "/test/home/git-bundle-server/www/test/myrepo/",
		},
		[]string{
			`[bundle]`,
			`	version = 1`,
			`	mode = all`,
			`	heuristic = creationToken`,
			``,
			`[bundle "1"]`,
			`	uri = bundle-1.bundle`,
			`	creationToken = 1`,
			``,
		},
		[]string{
			`[bundle]`,
			`	version = 1`,
			`	mode = all`,
			`	heuristic = creationToken`,
			``,
			`[bundle "1"]`,
			`	uri = myrepo/bundle-1.bundle`,
			`	creationToken = 1`,
			``,
		},
		false,
	},
	{
		"Multi-bundle list is sorted by creationToken",
		&bundles.BundleList{
			Version:   1,
			Mode:      "all",
			Heuristic: "creationToken",
			Bundles: map[int64]bundles.Bundle{
				2: {
					URI:           "/test/myrepo/bundle-2.bundle",
					Filename:      "/test/home/git-bundle-server/www/test/myrepo/bundle-2.bundle",
					CreationToken: 2,
				},
				5: {
					URI:           "/test/myrepo/bundle-5.bundle",
					Filename:      "/test/home/git-bundle-server/www/test/myrepo/bundle-5.bundle",
					CreationToken: 5,
				},
				1: {
					URI:           "/test/myrepo/bundle-1.bundle",
					Filename:      "/test/home/git-bundle-server/www/test/myrepo/bundle-1.bundle",
					CreationToken: 1,
				},
			},
		},
		&core.Repository{
			Route:   "test/myrepo",
			RepoDir: "/test/home/git-bundle-server/git/test/myrepo/",
			WebDir:  "/test/home/git-bundle-server/www/test/myrepo/",
		},
		[]string{
			`[bundle]`,
			`	version = 1`,
			`	mode = all`,
			`	heuristic = creationToken`,
			``,
			`[bundle "1"]`,
			`	uri = bundle-1.bundle`,
			`	creationToken = 1`,
			``,
			`[bundle "2"]`,
			`	uri = bundle-2.bundle`,
			`	creationToken = 2`,
			``,
			`[bundle "5"]`,
			`	uri = bundle-5.bundle`,
			`	creationToken = 5`,
			``,
		},
		[]string{
			`[bundle]`,
			`	version = 1`,
			`	mode = all`,
			`	heuristic = creationToken`,
			``,
			`[bundle "1"]`,
			`	uri = myrepo/bundle-1.bundle`,
			`	creationToken = 1`,
			``,
			`[bundle "2"]`,
			`	uri = myrepo/bundle-2.bundle`,
			`	creationToken = 2`,
			``,
			`[bundle "5"]`,
			`	uri = myrepo/bundle-5.bundle`,
			`	creationToken = 5`,
			``,
		},
		false,
	},
}

func TestBundles_WriteBundleList(t *testing.T) {
	testLogger := &MockTraceLogger{}
	testFileSystem := &MockFileSystem{}

	bundleListLockFile := &MockLockFile{}
	bundleListLockFile.On("Commit").Return(nil)
	repoBundleListLockFile := &MockLockFile{}
	repoBundleListLockFile.On("Commit").Return(nil)

	var mockWriteFunc func(io.Writer) error
	var writeErr error

	bundleProvider := bundles.NewBundleProvider(testLogger, testFileSystem, nil)
	for _, tt := range writeBundleListTests {
		t.Run(tt.title, func(t *testing.T) {
			// Set up mocks
			bundleListBuf := &bytes.Buffer{}
			testFileSystem.On("WriteLockFileFunc",
				filepath.Join(tt.repo.WebDir, bundles.BundleListFilename),
				mock.MatchedBy(func(writeFunc func(io.Writer) error) bool {
					mockWriteFunc = writeFunc
					return true
				}),
			).Run(
				func(mock.Arguments) { writeErr = mockWriteFunc(bundleListBuf) },
			).Return(bundleListLockFile, writeErr).Once()

			repoBundleListBuf := &bytes.Buffer{}
			testFileSystem.On("WriteLockFileFunc",
				filepath.Join(tt.repo.WebDir, bundles.RepoBundleListFilename),
				mock.MatchedBy(func(writeFunc func(io.Writer) error) bool {
					mockWriteFunc = writeFunc
					return true
				}),
			).Run(
				func(mock.Arguments) { writeErr = mockWriteFunc(repoBundleListBuf) },
			).Return(repoBundleListLockFile, writeErr).Once()

			jsonLockFile := &MockLockFile{}
			jsonLockFile.On("Commit").Return(nil).Once()
			testFileSystem.On("WriteLockFileFunc",
				filepath.Join(tt.repo.RepoDir, bundles.BundleListJsonFilename),
				mock.Anything,
			).Return(jsonLockFile, nil)

			// Run 'WriteBundleList()'
			err := bundleProvider.WriteBundleList(context.Background(), tt.bundleList, tt.repo)

			// Assert on expected values
			if tt.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			actualBundleList := bundleListBuf.String()
			expectedBundleList := ConcatLines(tt.bundleListFile)
			assert.Equal(t, expectedBundleList, actualBundleList)

			actualRepoBundleList := repoBundleListBuf.String()
			expectedRepoBundleList := ConcatLines(tt.repoBundleListFile)
			assert.Equal(t, expectedRepoBundleList, actualRepoBundleList)

			// Reset mocks
			testFileSystem.Mock = mock.Mock{}
		})
	}
}
