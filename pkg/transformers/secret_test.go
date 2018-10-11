package transformers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
)

type secretTransformerArgs struct {
	config    *types.Kustomization
	resources resmap.ResMap
}

func TestSecretRun(t *testing.T) {
	var secret = gvk.Gvk{Version: "v1", Kind: "Secret"}

	cert, err := ioutil.ReadFile("./testdata/tls.cert")
	if err != nil {
		t.Fatalf("Couldn't load tls.cert as test data")
	}
	key, err := ioutil.ReadFile("./testdata/tls.key")
	if err != nil {
		t.Fatalf("Couldn't load tls.key as test data")
	}

	for _, test := range []struct {
		name     string
		input    *secretTransformerArgs
		expected *secretTransformerArgs
	}{
		{
			name: "it should retrieve secrets",
			input: &secretTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(secret, "secret1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name": "secret1",
							},
							"type": string(corev1.SecretTypeOpaque),
							"data": map[string]interface{}{
								"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
								"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("password")),
							},
						}),
					resource.NewResId(secret, "secret2"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name": "secret2",
							},
							"type": string(corev1.SecretTypeTLS),
							"data": map[string]interface{}{
								"tls.cert": base64.StdEncoding.EncodeToString(cert),
								"tls.key":  base64.StdEncoding.EncodeToString(key),
							},
						}),
				},
			},
			expected: &secretTransformerArgs{
				config: &types.Kustomization{
					SecretGenerator: []types.SecretArgs{
						types.SecretArgs{
							Name: "secret1",
							CommandSources: types.CommandSources{
								Commands: map[string]string{
									"DB_USERNAME": "printf \\\"admin\\\"",
									"DB_PASSWORD": "printf \\\"password\\\"",
								},
							},
							Type: string(corev1.SecretTypeOpaque),
						},
						types.SecretArgs{
							Name: "secret2",
							CommandSources: types.CommandSources{
								Commands: map[string]string{
									"tls.cert": "printf \\\"" + string(cert) + "\\\"",
									"tls.key":  "printf \\\"" + string(key) + "\\\"",
								},
							},
							Type: string(corev1.SecretTypeTLS),
						},
					},
				},
				resources: resmap.ResMap{},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewSecretTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(test.input.config.SecretGenerator, test.expected.config.SecretGenerator) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config.SecretGenerator),
					spew.Sdump(test.input.config.SecretGenerator),
				)
			}

			if !reflect.DeepEqual(test.input.resources, test.expected.resources) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.resources),
					spew.Sdump(test.input.resources),
				)
			}
		})
	}
}