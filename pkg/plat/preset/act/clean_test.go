// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/rwtodd/Go.Sed/sed"
	"golang.org/x/sys/unix"

	"github.com/defended-net/malwatch/pkg/boot/env"
	"github.com/defended-net/malwatch/pkg/boot/env/cfg/act"
	"github.com/defended-net/malwatch/pkg/db/orm/hit"
	"github.com/defended-net/malwatch/pkg/fsys"
	"github.com/defended-net/malwatch/pkg/plat/acter"
	"github.com/defended-net/malwatch/pkg/scan/state"
	"github.com/defended-net/malwatch/pkg/sig"
)

var (
	reB64 = []string{
		`s/<?.*eval\(base64_decode\(.*?>//`,
		`s/<?php.*eval\(base64_decode\(.*?>//`,
		`s/eval\(base64_decode\([^;]*;//`,
	}

	reGz = []string{
		`s/<?.*eval\(gzinflate\(base64_decode\(.*?>//`,
		`s/<?php.*eval\(gzinflate\(base64_decode\(.*?>//`,
		`s/eval\(gzinflate\(base64_decode\(.*\);//`,
	}
)

func TestNewCleaner(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	var (
		got = NewCleaner(env)

		want = &Cleaner{
			verb:  VerbClean,
			dir:   env.Cfg.Acts.Quarantine.Dir,
			blkSz: got.blkSz,
			expr:  env.Cfg.Acts.Clean,
			rules: env.Paths.Sigs.Yrc,
		}
	)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cleaner result %v, want %v", got, want)
	}
}

func TestCleanLoad(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %s", err)
	}

	acter := &Cleaner{
		dir:   t.TempDir(),
		rules: env.Paths.Sigs.Yrc,
	}

	if err := acter.Load(); err != nil {
		t.Errorf("cleaner load error: %v", err)
	}
}

func TestCleanDisabled(t *testing.T) {
	input := &Cleaner{}

	if got := input.Load(); !errors.Is(got, acter.ErrDisabled) {
		t.Errorf("unexpected cleaner load error %v, want %v", got, acter.ErrDisabled)
	}
}

func TestCleanVerb(t *testing.T) {
	acter := &Cleaner{
		verb: VerbClean,
	}

	if got := acter.Verb(); got != VerbClean {
		t.Errorf("unexpected verb result %v, want %v", got, VerbClean)
	}
}

func TestCleanInject(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %v", err)
	}

	acter := NewCleaner(env)

	if err := acter.Load(); err != nil {
		t.Fatalf("cleaner load error: %s", err)
	}

	acter.expr = act.Clean{
		"gz": reGz,
	}

	malware := []byte(`<?php echo "hello world";
eval(gzinflate(base64_decode('test')));
?>`)

	path := filepath.Join(t.TempDir(), t.Name())

	if err := os.WriteFile(path, []byte(malware), 0660); err != nil {
		t.Fatalf("file write error: %v", err)
	}

	stat := &unix.Stat_t{}

	if err := unix.Stat(path, stat); err != nil {
		t.Fatalf("stat error: %v", err)
	}

	input := state.NewResult("",
		state.Paths{
			path: hit.NewMeta(
				fsys.NewAttr(stat),

				[]string{"gz"},

				"clean",
			),
		})

	if err := acter.Act(input); err != nil {
		t.Errorf("clean act error: %v", err)
	}
}

func TestCleanInjectMultiLine(t *testing.T) {
	env, err := env.Mock(t.Name(), t.TempDir())
	if err != nil {
		t.Fatalf("env mock error: %s", err)
	}

	if err := sig.Mock(env); err != nil {
		t.Fatalf("sig mock error: %s", err)
	}

	acter := &Cleaner{
		dir:   t.TempDir(),
		rules: env.Paths.Sigs.Yrc,

		blkSz: 32768,

		expr: map[string][]string{
			"php_base64_inject": {
				`s/<?.*eval\(base64_decode\(.*?>//`,
				`s/<?php.*eval\(base64_decode\(.*?>//`,
				`s/eval\(base64_decode\([^;]*;//`,
			},
		},
	}

	if err := acter.Load(); err != nil {
		t.Fatalf("cleaner load error: %v", err)
	}

	var (
		path = filepath.Join(t.TempDir(), t.Name())

		sample = `<?php echo "hello world";
eval(base64_decode("mal"));
echo "foo";
eval(base64_decode("ware"));
?>`
	)

	if err := os.WriteFile(path, []byte(sample), 0600); err != nil {
		t.Fatalf("hit file write error: %s", err)
	}

	if err := acter.clean(path, &hit.Meta{
		Rules: []string{"php_base64_inject"},

		Attr: &fsys.Attr{
			UID:  os.Getuid(),
			GID:  os.Getgid(),
			Mode: 0600,
		},
	}); err != nil {
		t.Errorf("clean error: %s", err)
	}
}

func TestCleanB64(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"match": {
			input: `<?php echo "hello world";
eval(base64_decode("test"));
?>`,
			want: `<?php echo "hello world";

?>
`,
		},

		"no-match": {
			input: `<?php echo "hello world";?>`,
			want: `<?php echo "hello world";?>
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, re := range reB64 {
				sed, err := sed.New(strings.NewReader(re))
				if err != nil {
					t.Fatalf("sed init error: %v", err)
				}

				test.input, err = sed.RunString(test.input)
				if err != nil {
					t.Fatalf("sed run error: %v", err)
				}
			}

			if test.input != test.want {
				t.Errorf("unexpected clean result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestCleanB64Multi(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"match": {
			input: `<?php echo "hello world";
eval(base64_decode("mal"));
echo "foo";
eval(base64_decode("ware"));
?>`,

			want: `<?php echo "hello world";

echo "foo";

?>
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, re := range reB64 {
				sed, err := sed.New(strings.NewReader(re))
				if err != nil {
					t.Fatalf("sed init error: %v", err)
				}

				test.input, err = sed.RunString(test.input)
				if err != nil {
					t.Fatalf("sed run error: %v", err)
				}
			}

			if test.input != test.want {
				t.Errorf("unexpected clean result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestCleanGzB64(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"match": {
			input: `<?php echo "hello world";
eval(gzinflate(base64_decode('test')));
?>`,

			want: `<?php echo "hello world";

?>
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, re := range reGz {
				sed, err := sed.New(strings.NewReader(re))
				if err != nil {
					t.Fatalf("sed init error: %v", err)
				}

				test.input, err = sed.RunString(test.input)
				if err != nil {
					t.Fatalf("sed run error: %v", err)
				}
			}

			if test.input != test.want {
				t.Errorf("unexpected clean result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestCleanMultiGzB64(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"match": {
			input: `<?php echo "hello world";
eval(gzinflate(base64_decode('test')));
echo "foo";
eval(gzinflate(base64_decode('test')));
?>`,

			want: `<?php echo "hello world";

echo "foo";

?>
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var result string

			for _, re := range append(reB64, reGz...) {
				sed, err := sed.New(strings.NewReader(re))
				if err != nil {
					t.Fatalf("sed init error: %v", err)
				}

				result, err = sed.RunString(string(test.input))
				if err != nil {
					t.Fatalf("sed run error: %v", err)
				}
			}

			if result != test.want {
				t.Errorf("unexpected clean result %v, want %v", result, test.want)
			}
		})
	}
}

func TestCleanComposite(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"match": {
			input: `<?php echo "hello world";
eval(base64_decode("test"));
echo "foo";
eval(gzinflate(base64_decode('test')));
?>`,

			want: `<?php echo "hello world";

echo "foo";

?>
`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, re := range append(reB64, reGz...) {
				sed, err := sed.New(strings.NewReader(re))
				if err != nil {
					t.Fatalf("sed init error: %v", err)
				}

				test.input, err = sed.RunString(string(test.input))
				if err != nil {
					t.Fatalf("sed run error: %v", err)
				}
			}

			if test.input != test.want {
				t.Errorf("unexpected clean result %v, want %v", test.input, test.want)
			}
		})
	}
}

func TestCleanErrs(t *testing.T) {
	var (
		input = &Cleaner{}
		want  = ErrQuarantineNoDir
	)

	if err := input.Act(nil); !errors.Is(err, want) {
		t.Errorf("unexpected clean error %v, want %v", input, want)
	}
}
