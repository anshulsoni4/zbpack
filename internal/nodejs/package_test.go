package nodejs_test

import (
	"encoding/json"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/nodejs"
)

func TestNewPackageJson(t *testing.T) {
	p := nodejs.NewPackageJSON()

	// dependencies
	assert.NotPanics(t, func() {
		_, ok := p.Dependencies["astro"]
		assert.False(t, ok)
	})

	// devDependencies
	assert.NotPanics(t, func() {
		_, ok := p.DevDependencies["astro"]
		assert.False(t, ok)
	})

	// scripts
	assert.NotPanics(t, func() {
		_, ok := p.Scripts["build"]
		assert.False(t, ok)
	})

	// panic!
	assert.Panics(t, func() {
		p.Scripts["build"] = "hi"
	})

	// engines
	assert.Empty(t, p.Engines.Node)

	// main
	assert.Empty(t, p.Main)
}

func TestDeserializePackageJson_NoSuchFile(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, err := nodejs.DeserializePackageJSON(fs)
	assert.Error(t, err)
}

func TestDeserializePackageJson_InvalidFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte("\\0"), 0o644)

	_, err := nodejs.DeserializePackageJSON(fs)
	assert.Error(t, err)
}

func TestDeserializePackageJson_EmptyJson(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte("{}"), 0o644)

	p, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)
	assert.Nil(t, p.Dependencies)
	assert.Nil(t, p.DevDependencies)
	assert.Nil(t, p.Scripts)
	assert.Empty(t, p.Engines.Node)
	assert.Empty(t, p.Main)
}

func TestDeserializePackageJson_NoMatchField(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "package.json", []byte(`{"this_is_a_test": "hhh"}`), 0o644)

	_, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)
}

func TestDeserializePackageJson_WithDepsAndEngines(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{
		"dependencies": map[string]string{
			"astro": "0.0.1",
		},
		"devDependencies": map[string]string{
			"prettier": "^1.2.3",
		},
		"engines": map[string]string{
			"node": "^18",
		},
	})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	version, ok := packageJSON.Dependencies["astro"]
	assert.True(t, ok)
	assert.Equal(t, "0.0.1", version)

	version, ok = packageJSON.DevDependencies["prettier"]
	assert.True(t, ok)
	assert.Equal(t, "^1.2.3", version)

	assert.Equal(t, "^18", packageJSON.Engines.Node)
}

func TestDeserializePackageJson_WithMainAndEngines(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{
		"main": "hello",
		"engines": map[string]string{
			"node": "^18",
		},
	})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	assert.Equal(t, "hello", packageJSON.Main)
	assert.Equal(t, "^18", packageJSON.Engines.Node)
}

func TestDeserializePackageJson_WithPackageManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{
		"packageManager": "yarn@1.2.3",
	})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	assert.NotNil(t, packageJSON.PackageManager)
	assert.Equal(t, "yarn@1.2.3", *packageJSON.PackageManager)
}

func TestDeserializePackageJson_WithoutPackageManager(t *testing.T) {
	fs := afero.NewMemMapFs()
	data, err := json.Marshal(map[string]interface{}{})
	assert.NoError(t, err)

	_ = afero.WriteFile(fs, "package.json", data, 0o644)

	packageJSON, err := nodejs.DeserializePackageJSON(fs)
	assert.NoError(t, err)

	assert.Nil(t, packageJSON.PackageManager)
}
