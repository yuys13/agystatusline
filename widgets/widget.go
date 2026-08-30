package widgets

import (
	"bytes"
	"os"
	"os/exec"
	"sync"

	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/utils"
)

// Widget defines the interface that all statusline widgets must implement.
type Widget interface {
	GetDefaultColor() string
	GetDisplayName() string
	Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (title string, body string, err error)
	GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string
}

type CwdResolver = utils.CwdResolver

var (
	registry = make(map[string]Widget)
	regMutex sync.RWMutex
)

// RegisterWidget registers a widget with the global registry.
func RegisterWidget(name string, w Widget) {
	regMutex.Lock()
	defer regMutex.Unlock()
	registry[name] = w
}

// GetWidget retrieves a widget by name from the registry.
func GetWidget(name string) Widget {
	regMutex.RLock()
	defer regMutex.RUnlock()
	return registry[name]
}

// runGitExec runs a real git command on the OS.
func runGitExec(args []string, cwd string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

// runGitCommand wraps utils.RunGit to execute commands with cache.
var runGitCommand = func(command string, ctx CwdResolver, ttlSeconds int) (string, error) {
	return utils.RunGit(command, ctx, ttlSeconds, runGitExec)
}

// RegisterAll registers all available widget implementations.
func RegisterAll() {
	RegisterWidget("model", &ModelWidget{})
	RegisterWidget("git-branch", &GitBranchWidget{})
	RegisterWidget("git-changes", &GitChangesWidget{})
	RegisterWidget("custom-text", &CustomTextWidget{})
	RegisterWidget("sandbox", &SandboxWidget{})
	RegisterWidget("agent-state", &AgentStateWidget{})
	RegisterWidget("context-bar", &ContextBarWidget{})
	RegisterWidget("artifacts", &ArtifactsWidget{})
	RegisterWidget("subagents", &SubagentsWidget{})
	RegisterWidget("tasks", &TasksWidget{})

	// Quota widgets
	RegisterWidget("quota", &QuotaWidget{DisplayName: "Quota"})
	RegisterWidget("quota-5h", &QuotaWidget{DefaultKey: "gemini-5h", DefaultTitle: "5h", DisplayName: "Quota: 5h"})
	RegisterWidget("quota-7d", &QuotaWidget{DefaultKey: "gemini-weekly", DefaultTitle: "7d", DisplayName: "Quota: 7d"})
	RegisterWidget("quota-3p-5h", &QuotaWidget{DefaultKey: "3p-5h", DefaultTitle: "3p-5h", DisplayName: "Quota: 3P 5h"})
	RegisterWidget("quota-3p-7d", &QuotaWidget{DefaultKey: "3p-weekly", DefaultTitle: "3p-7d", DisplayName: "Quota: 3P 7d"})

	// Quota Bar widgets
	RegisterWidget("quota-bar", &QuotaBarWidget{DisplayName: "Quota Bar"})
	RegisterWidget("quota-bar-5h", &QuotaBarWidget{DefaultKey: "gemini-5h", DefaultTitle: "5h", DisplayName: "Quota Bar: 5h"})
	RegisterWidget("quota-bar-7d", &QuotaBarWidget{DefaultKey: "gemini-weekly", DefaultTitle: "7d", DisplayName: "Quota Bar: 7d"})
	RegisterWidget("quota-bar-3p-5h", &QuotaBarWidget{DefaultKey: "3p-5h", DefaultTitle: "3p-5h", DisplayName: "Quota Bar: 3P 5h"})
	RegisterWidget("quota-bar-3p-7d", &QuotaBarWidget{DefaultKey: "3p-weekly", DefaultTitle: "3p-7d", DisplayName: "Quota Bar: 3P 7d"})
}
