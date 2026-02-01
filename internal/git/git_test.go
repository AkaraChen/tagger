package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "tagger-git-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	cmd.Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = tmpDir
	cmd.Run()

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewGitClient(t *testing.T) {
	client := NewGitClient(".")
	if client == nil {
		t.Fatal("NewGitClient returned nil")
	}
	if client.workDir != "." {
		t.Errorf("expected workDir '.', got '%s'", client.workDir)
	}
}

func TestIsGitRepository(t *testing.T) {
	t.Run("valid git repo", func(t *testing.T) {
		tmpDir, cleanup := setupTestRepo(t)
		defer cleanup()

		client := NewGitClient(tmpDir)
		isRepo, err := client.IsGitRepository()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !isRepo {
			t.Error("expected true for git repository")
		}
	})

	t.Run("not a git repo", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "tagger-test-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		client := NewGitClient(tmpDir)
		isRepo, err := client.IsGitRepository()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if isRepo {
			t.Error("expected false for non-git directory")
		}
	})
}

func TestHasUncommittedChanges(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("no changes", func(t *testing.T) {
		hasChanges, err := client.HasUncommittedChanges()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if hasChanges {
			t.Error("expected no uncommitted changes")
		}
	})

	t.Run("with changes", func(t *testing.T) {
		// Create a new file
		os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("new"), 0644)

		hasChanges, err := client.HasUncommittedChanges()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !hasChanges {
			t.Error("expected uncommitted changes")
		}
	})
}

func TestGetAllTags(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("no tags", func(t *testing.T) {
		tags, err := client.GetAllTags()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(tags) != 0 {
			t.Errorf("expected 0 tags, got %d", len(tags))
		}
	})

	t.Run("with tags", func(t *testing.T) {
		// Create some tags
		cmd := exec.Command("git", "tag", "v1.0.0")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "tag", "v1.1.0")
		cmd.Dir = tmpDir
		cmd.Run()

		tags, err := client.GetAllTags()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tags))
		}
	})
}

func TestTagExists(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	// Create a tag
	cmd := exec.Command("git", "tag", "v1.0.0")
	cmd.Dir = tmpDir
	cmd.Run()

	t.Run("existing tag", func(t *testing.T) {
		exists, err := client.TagExists("v1.0.0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !exists {
			t.Error("expected tag to exist")
		}
	})

	t.Run("non-existing tag", func(t *testing.T) {
		exists, err := client.TagExists("v2.0.0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if exists {
			t.Error("expected tag to not exist")
		}
	})
}

func TestCreateTag(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("create lightweight tag", func(t *testing.T) {
		err := client.CreateTag("v1.0.0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		exists, _ := client.TagExists("v1.0.0")
		if !exists {
			t.Error("tag was not created")
		}
	})

	t.Run("create duplicate tag", func(t *testing.T) {
		err := client.CreateTag("v1.0.0")
		if err == nil {
			t.Error("expected error when creating duplicate tag")
		}
	})
}

func TestCreateAnnotatedTag(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("create annotated tag", func(t *testing.T) {
		err := client.CreateAnnotatedTag("v2.0.0", "Release v2.0.0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		exists, _ := client.TagExists("v2.0.0")
		if !exists {
			t.Error("annotated tag was not created")
		}
	})
}

func TestHasRemote(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("no remote", func(t *testing.T) {
		hasRemote, err := client.HasRemote()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if hasRemote {
			t.Error("expected no remote")
		}
	})

	t.Run("with remote", func(t *testing.T) {
		// Add a remote
		cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
		cmd.Dir = tmpDir
		cmd.Run()

		hasRemote, err := client.HasRemote()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !hasRemote {
			t.Error("expected remote to exist")
		}
	})
}

func TestGetRemoteName(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("no remote", func(t *testing.T) {
		_, err := client.GetRemoteName()
		if err == nil {
			t.Error("expected error when no remote exists")
		}
	})

	t.Run("with origin", func(t *testing.T) {
		cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
		cmd.Dir = tmpDir
		cmd.Run()

		name, err := client.GetRemoteName()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if name != "origin" {
			t.Errorf("expected 'origin', got '%s'", name)
		}
	})

	t.Run("prefers origin", func(t *testing.T) {
		// Add another remote
		cmd := exec.Command("git", "remote", "add", "upstream", "https://github.com/other/repo.git")
		cmd.Dir = tmpDir
		cmd.Run()

		name, err := client.GetRemoteName()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if name != "origin" {
			t.Errorf("expected 'origin', got '%s'", name)
		}
	})
}

func TestGetRemoteURL(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("https url", func(t *testing.T) {
		cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
		cmd.Dir = tmpDir
		cmd.Run()

		url, err := client.GetRemoteURL()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if url != "https://github.com/test/repo" {
			t.Errorf("expected 'https://github.com/test/repo', got '%s'", url)
		}

		// Clean up for next test
		cmd = exec.Command("git", "remote", "remove", "origin")
		cmd.Dir = tmpDir
		cmd.Run()
	})

	t.Run("ssh url conversion", func(t *testing.T) {
		cmd := exec.Command("git", "remote", "add", "origin", "git@github.com:test/repo.git")
		cmd.Dir = tmpDir
		cmd.Run()

		url, err := client.GetRemoteURL()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if url != "https://github.com/test/repo" {
			t.Errorf("expected 'https://github.com/test/repo', got '%s'", url)
		}
	})
}

func TestGetTagsWithDates(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	client := NewGitClient(tmpDir)

	t.Run("no tags", func(t *testing.T) {
		tags, err := client.GetTagsWithDates()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(tags) != 0 {
			t.Errorf("expected 0 tags, got %d", len(tags))
		}
	})

	t.Run("with tags", func(t *testing.T) {
		// Create tags
		cmd := exec.Command("git", "tag", "v1.0.0")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "tag", "-a", "v1.1.0", "-m", "release")
		cmd.Dir = tmpDir
		cmd.Run()

		tags, err := client.GetTagsWithDates()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tags))
		}

		// Check that tags have names
		for _, tag := range tags {
			if tag.Name == "" {
				t.Error("expected tag name to be non-empty")
			}
		}
	})
}
