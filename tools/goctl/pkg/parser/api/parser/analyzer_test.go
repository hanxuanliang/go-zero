package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/assertx"
)

func Test_Parse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		apiSpec, err := Parse("./testdata/example-dev.api", nil)
		assert.Nil(t, err)

		specStr, _ := json.Marshal(apiSpec)
		println(string(specStr))
	})
	t.Run("invalid", func(t *testing.T) {
		data, err := os.ReadFile("./testdata/invalid.api")
		assert.NoError(t, err)
		splits := bytes.Split(data, []byte("-----"))
		var testFile []string
		for idx, split := range splits {
			replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "", "\f", "")
			r := replacer.Replace(string(split))
			if len(r) == 0 {
				continue
			}
			filename := filepath.Join(t.TempDir(), fmt.Sprintf("invalid%d.api", idx))
			err := os.WriteFile(filename, split, 0666)
			assert.NoError(t, err)
			testFile = append(testFile, filename)
		}
		for _, v := range testFile {
			_, err := Parse(v, nil)
			assertx.Error(t, err)
		}
	})
	t.Run("circleImport", func(t *testing.T) {
		_, err := Parse("./testdata/base.api", nil)
		assertx.Error(t, err)
	})
	t.Run("link_import", func(t *testing.T) {
		_, err := Parse("./testdata/link_import.api", nil)
		assert.Nil(t, err)
	})
}
