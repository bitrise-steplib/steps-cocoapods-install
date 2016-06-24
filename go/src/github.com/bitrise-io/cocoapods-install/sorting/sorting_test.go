package sorting

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathDept(t *testing.T) {
	t.Log("empty path - error")
	{
		actual, err := PathDepth("")
		require.Error(t, err)
		require.Equal(t, 0, actual)
	}

	t.Log("simple path")
	{
		actual, err := PathDepth("/Podfile")
		require.NoError(t, err)
		require.Equal(t, 1, actual)
	}

	t.Log("simple path")
	{
		actual, err := PathDepth("/dir/Podfile")
		require.NoError(t, err)
		require.Equal(t, 2, actual)
	}

	t.Log("relative path")
	{
		currentDir, err := os.Getwd()
		currentDirDepth, err := PathDepth(currentDir)
		require.NoError(t, err)

		actual, err := PathDepth("./Podfile")
		require.NoError(t, err)
		require.Equal(t, currentDirDepth+1, actual)
	}
}

func TestByComponents(t *testing.T) {
	t.Log("Simple sort")
	{
		fileList := []string{
			"path/to",
			"path/to/my",
			"path",
		}

		sort.Sort(ByComponents(fileList))
		require.Equal(t, []string{"path", "path/to", "path/to/my"}, fileList)
	}

	t.Log("Path with equal components length - alpahabetic sort")
	{
		fileList := []string{
			"path1",
			"path/to",
			"path/to/my",
			"path",
		}

		sort.Sort(ByComponents(fileList))
		require.Equal(t, 4, len(fileList))
		require.Equal(t, "path", fileList[0])
		require.Equal(t, "path1", fileList[1])
		require.Equal(t, "path/to", fileList[2])
		require.Equal(t, "path/to/my", fileList[3])
	}

	t.Log("Relative path with equal components length - alpahabetic sort")
	{
		fileList := []string{
			"./c",
			"./a",
			"b",
		}

		sort.Sort(ByComponents(fileList))
		require.Equal(t, 3, len(fileList))
		require.Equal(t, "./a", fileList[0])
		require.Equal(t, "b", fileList[1])
		require.Equal(t, "./c", fileList[2])
	}
}
