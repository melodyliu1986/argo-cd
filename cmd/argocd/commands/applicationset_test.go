package commands

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
)

func TestPrintApplicationSetNames(t *testing.T) {
	output, _ := captureOutput(func() error {
		appSet := &v1alpha1.ApplicationSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		}
		appSet2 := &v1alpha1.ApplicationSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "team-one",
				Name:      "test",
			},
		}
		printApplicationSetNames([]v1alpha1.ApplicationSet{*appSet, *appSet2})
		return nil
	})
	expectation := "test\nteam-one/test\n"
	require.Equalf(t, output, expectation, "Incorrect print params output %q, should be %q", output, expectation)
}

func TestPrintApplicationSetTable(t *testing.T) {
	output, err := captureOutput(func() error {
		app := &v1alpha1.ApplicationSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "app-name",
			},
			Spec: v1alpha1.ApplicationSetSpec{
				Generators: []v1alpha1.ApplicationSetGenerator{
					{
						Git: &v1alpha1.GitGenerator{
							RepoURL:  "https://github.com/argoproj/argo-cd.git",
							Revision: "head",
							Directories: []v1alpha1.GitDirectoryGeneratorItem{
								{
									Path: "applicationset/examples/git-generator-directory/cluster-addons/*",
								},
							},
						},
					},
				},
				Template: v1alpha1.ApplicationSetTemplate{
					Spec: v1alpha1.ApplicationSpec{
						Project: "default",
					},
				},
			},
			Status: v1alpha1.ApplicationSetStatus{
				Conditions: []v1alpha1.ApplicationSetCondition{
					{
						Status: v1alpha1.ApplicationSetConditionStatusTrue,
						Type:   v1alpha1.ApplicationSetConditionResourcesUpToDate,
					},
				},
			},
		}

		app2 := &v1alpha1.ApplicationSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-name",
				Namespace: "team-two",
			},
			Spec: v1alpha1.ApplicationSetSpec{
				Generators: []v1alpha1.ApplicationSetGenerator{
					{
						Git: &v1alpha1.GitGenerator{
							RepoURL:  "https://github.com/argoproj/argo-cd.git",
							Revision: "head",
							Directories: []v1alpha1.GitDirectoryGeneratorItem{
								{
									Path: "applicationset/examples/git-generator-directory/cluster-addons/*",
								},
							},
						},
					},
				},
				Template: v1alpha1.ApplicationSetTemplate{
					Spec: v1alpha1.ApplicationSpec{
						Project: "default",
					},
				},
			},
			Status: v1alpha1.ApplicationSetStatus{
				Conditions: []v1alpha1.ApplicationSetCondition{
					{
						Status: v1alpha1.ApplicationSetConditionStatusTrue,
						Type:   v1alpha1.ApplicationSetConditionResourcesUpToDate,
					},
				},
			},
		}
		output := "table"
		printApplicationSetTable([]v1alpha1.ApplicationSet{*app, *app2}, &output)
		return nil
	})
	require.NoError(t, err)
	expectation := "NAME               PROJECT  SYNCPOLICY  CONDITIONS\napp-name           default  nil         [{ResourcesUpToDate  <nil> True }]\nteam-two/app-name  default  nil         [{ResourcesUpToDate  <nil> True }]\n"
	assert.Equal(t, expectation, output)
}

func TestPrintAppSetSummaryTable(t *testing.T) {
	baseAppSet := &v1alpha1.ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-name",
		},
		Spec: v1alpha1.ApplicationSetSpec{
			Generators: []v1alpha1.ApplicationSetGenerator{
				{
					Git: &v1alpha1.GitGenerator{
						RepoURL:  "https://github.com/argoproj/argo-cd.git",
						Revision: "head",
						Directories: []v1alpha1.GitDirectoryGeneratorItem{
							{
								Path: "applicationset/examples/git-generator-directory/cluster-addons/*",
							},
						},
					},
				},
			},
			Template: v1alpha1.ApplicationSetTemplate{
				Spec: v1alpha1.ApplicationSpec{
					Project: "default",
				},
			},
		},
		Status: v1alpha1.ApplicationSetStatus{
			Conditions: []v1alpha1.ApplicationSetCondition{
				{
					Status: v1alpha1.ApplicationSetConditionStatusTrue,
					Type:   v1alpha1.ApplicationSetConditionResourcesUpToDate,
				},
			},
		},
	}
	appsetSpecSource := baseAppSet.DeepCopy()
	appsetSpecSource.Spec.Template.Spec.Source = &v1alpha1.ApplicationSource{
		RepoURL:        "test1",
		TargetRevision: "master1",
		Path:           "/test1",
	}

	appsetSpecSources := baseAppSet.DeepCopy()
	appsetSpecSources.Spec.Template.Spec.Sources = v1alpha1.ApplicationSources{
		{
			RepoURL:        "test1",
			TargetRevision: "master1",
			Path:           "/test1",
		},
		{
			RepoURL:        "test2",
			TargetRevision: "master2",
			Path:           "/test2",
		},
	}

	appsetSpecSyncPolicy := baseAppSet.DeepCopy()
	appsetSpecSyncPolicy.Spec.SyncPolicy = &v1alpha1.ApplicationSetSyncPolicy{
		PreserveResourcesOnDeletion: true,
	}

	appSetTemplateSpecSyncPolicy := baseAppSet.DeepCopy()
	appSetTemplateSpecSyncPolicy.Spec.Template.Spec.SyncPolicy = &v1alpha1.SyncPolicy{
		Automated: &v1alpha1.SyncPolicyAutomated{
			SelfHeal: true,
		},
	}

	appSetBothSyncPolicies := baseAppSet.DeepCopy()
	appSetBothSyncPolicies.Spec.SyncPolicy = &v1alpha1.ApplicationSetSyncPolicy{
		PreserveResourcesOnDeletion: true,
	}
	appSetBothSyncPolicies.Spec.Template.Spec.SyncPolicy = &v1alpha1.SyncPolicy{
		Automated: &v1alpha1.SyncPolicyAutomated{
			SelfHeal: true,
		},
	}

	for _, tt := range []struct {
		name           string
		appSet         *v1alpha1.ApplicationSet
		expectedOutput string
	}{
		{
			name:   "appset with only spec.syncPolicy set",
			appSet: appsetSpecSyncPolicy,
			expectedOutput: `Name:               app-name
Project:            default
Server:             
Namespace:          
Source:
- Repo:             
  Target:           
SyncPolicy:         <none>
`,
		},
		{
			name:   "appset with only spec.template.spec.syncPolicy set",
			appSet: appSetTemplateSpecSyncPolicy,
			expectedOutput: `Name:               app-name
Project:            default
Server:             
Namespace:          
Source:
- Repo:             
  Target:           
SyncPolicy:         Automated
`,
		},
		{
			name:   "appset with both spec.SyncPolicy and spec.template.spec.syncPolicy set",
			appSet: appSetBothSyncPolicies,
			expectedOutput: `Name:               app-name
Project:            default
Server:             
Namespace:          
Source:
- Repo:             
  Target:           
SyncPolicy:         Automated
`,
		},
		{
			name:   "appset with a single source",
			appSet: appsetSpecSource,
			expectedOutput: `Name:               app-name
Project:            default
Server:             
Namespace:          
Source:
- Repo:             test1
  Target:           master1
  Path:             /test1
SyncPolicy:         <none>
`,
		},
		{
			name:   "appset with a multiple sources",
			appSet: appsetSpecSources,
			expectedOutput: `Name:               app-name
Project:            default
Server:             
Namespace:          
Sources:
- Repo:             test1
  Target:           master1
  Path:             /test1
- Repo:             test2
  Target:           master2
  Path:             /test2
SyncPolicy:         <none>
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			defer func() {
				os.Stdout = oldStdout
			}()

			r, w, _ := os.Pipe()
			os.Stdout = w

			printAppSetSummaryTable(tt.appSet)
			w.Close()

			out, err := io.ReadAll(r)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(out))
		})
	}
}
