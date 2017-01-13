package utility

import (
	"path/filepath"

	"fmt"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-xcode/xcodeproj"
)

const (
	embeddedWorkspacePathPattern = `.+\.xcodeproj/.+\.xcworkspace`

	gitDirName      = ".git"
	podsDirName     = "Pods"
	carthageDirName = "Carthage"

	frameworkExt = ".framework"
)

// AllowXcodeProjExtFilter ...
var AllowXcodeProjExtFilter = ExtensionFilter(xcodeproj.XCodeProjExt, true)

// AllowXCWorkspaceExtFilter ...
var AllowXCWorkspaceExtFilter = ExtensionFilter(xcodeproj.XCWorkspaceExt, true)

// AllowIsDirectoryFilter ...
var AllowIsDirectoryFilter = IsDirectoryFilter(true)

// ForbidEmbeddedWorkspaceRegexpFilter ...
var ForbidEmbeddedWorkspaceRegexpFilter = RegexpFilter(embeddedWorkspacePathPattern, false)

// ForbidGitDirComponentFilter ...
var ForbidGitDirComponentFilter = ComponentFilter(gitDirName, false)

// ForbidPodsDirComponentFilter ...
var ForbidPodsDirComponentFilter = ComponentFilter(podsDirName, false)

// ForbidCarthageDirComponentFilter ...
var ForbidCarthageDirComponentFilter = ComponentFilter(carthageDirName, false)

// ForbidFramworkComponentWithExtensionFilter ...
var ForbidFramworkComponentWithExtensionFilter = ComponentWithExtensionFilter(frameworkExt, false)

// AllowIphoneosSDKFilter ...
var AllowIphoneosSDKFilter = SDKFilter("iphoneos", true)

// AllowMacosxSDKFilter ...
var AllowMacosxSDKFilter = SDKFilter("macosx", true)

// SDKFilter ...
func SDKFilter(sdk string, allowed bool) FilterFunc {
	return func(pth string) (bool, error) {
		found := false

		projectFiles := []string{}

		if xcodeproj.IsXCodeProj(pth) {
			projectFiles = append(projectFiles, pth)
		} else if xcodeproj.IsXCWorkspace(pth) {
			projects, err := xcodeproj.WorkspaceProjectReferences(pth)
			if err != nil {
				return false, err
			}

			for _, project := range projects {
				exist, err := pathutil.IsPathExists(project)
				if err != nil {
					return false, err
				}
				if !exist {
					continue
				}
				projectFiles = append(projectFiles, project)

			}
		} else {
			return false, fmt.Errorf("Not Xcode project nor workspace file: %s", pth)
		}

		for _, projectFile := range projectFiles {
			pbxprojPth := filepath.Join(projectFile, "project.pbxproj")
			projectSDKs, err := xcodeproj.GetBuildConfigSDKs(pbxprojPth)
			if err != nil {
				return false, err
			}

			for _, projectSDK := range projectSDKs {
				if projectSDK == sdk {
					found = true
					break
				}
			}
		}

		return (allowed == found), nil
	}
}

// FindWorkspaceInList ...
func FindWorkspaceInList(workspacePth string, workspaces []xcodeproj.WorkspaceModel) (xcodeproj.WorkspaceModel, bool) {
	for _, workspace := range workspaces {
		if workspace.Pth == workspacePth {
			return workspace, true
		}
	}
	return xcodeproj.WorkspaceModel{}, false
}

// FindProjectInList ...
func FindProjectInList(projectPth string, projects []xcodeproj.ProjectModel) (xcodeproj.ProjectModel, bool) {
	for _, project := range projects {
		if project.Pth == projectPth {
			return project, true
		}
	}
	return xcodeproj.ProjectModel{}, false
}

// RemoveProjectFromList ...
func RemoveProjectFromList(projectPth string, projects []xcodeproj.ProjectModel) []xcodeproj.ProjectModel {
	newProjects := []xcodeproj.ProjectModel{}
	for _, project := range projects {
		if project.Pth != projectPth {
			newProjects = append(newProjects, project)
		}
	}
	return newProjects
}

// ReplaceWorkspaceInList ...
func ReplaceWorkspaceInList(workspaces []xcodeproj.WorkspaceModel, workspace xcodeproj.WorkspaceModel) []xcodeproj.WorkspaceModel {
	updatedWorkspaces := []xcodeproj.WorkspaceModel{}
	for _, w := range workspaces {
		if w.Pth == workspace.Pth {
			updatedWorkspaces = append(updatedWorkspaces, workspace)
		} else {
			updatedWorkspaces = append(updatedWorkspaces, w)
		}
	}
	return updatedWorkspaces
}

// CreateStandaloneProjectsAndWorkspaces ...
func CreateStandaloneProjectsAndWorkspaces(projectFiles, workspaceFiles []string) ([]xcodeproj.ProjectModel, []xcodeproj.WorkspaceModel, error) {
	workspaces := []xcodeproj.WorkspaceModel{}
	for _, workspaceFile := range workspaceFiles {
		workspace, err := xcodeproj.NewWorkspace(workspaceFile, projectFiles...)
		if err != nil {
			return []xcodeproj.ProjectModel{}, []xcodeproj.WorkspaceModel{}, err
		}
		workspaces = append(workspaces, workspace)
	}

	standaloneProjects := []xcodeproj.ProjectModel{}
	for _, projectFile := range projectFiles {
		workspaceContains := false
		for _, workspace := range workspaces {
			_, found := FindProjectInList(projectFile, workspace.Projects)
			if found {
				workspaceContains = true
				break
			}
		}

		if !workspaceContains {
			project, err := xcodeproj.NewProject(projectFile)
			if err != nil {
				return []xcodeproj.ProjectModel{}, []xcodeproj.WorkspaceModel{}, err
			}
			standaloneProjects = append(standaloneProjects, project)
		}
	}

	return standaloneProjects, workspaces, nil
}
