// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func isSafeExpr(expr string) error {
	trim := strings.TrimSpace(expr)
	if len(trim) == 0 {
		return fmt.Errorf("%w, %w", ErrCleanExprNotSafe, ErrCleanExprMissing)
	}

	toks, err := lexExpr(trim)
	if err != nil {
		return fmt.Errorf("%w, %w, %v %q", ErrCleanExprNotSafe, ErrCleanExprLex, err, expr)
	}

	if len(toks) == 0 {
		return fmt.Errorf("%w, %w, %q", ErrCleanExprNotSafe, ErrCleanExprMissingToks, expr)
	}

	for _, tok := range toks {
		switch tok.cmd {
		// safes
		case 's':
			if err := isSafeFlags(tok.flags, expr); err != nil {
				return err
			}

		case 'y':
		case 'd', 'D':
		case 'p', 'P':
		case 'g', 'G', 'h', 'H', 'x':
		case 'n', 'N':
		case 'b', 't':
		case '=':
		case 'q':
		case 'a', 'i', 'c':
		case ':':
		case ';':
		case '{', '}':
		case '!':
		case '$':
		case '/':
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':

		// unsafes
		case 'w', 'W':
			return fmt.Errorf("%w, %w, %q", ErrCleanExprNotSafe, ErrCleanSuppWrite, expr)

		case 'r', 'R':
			return fmt.Errorf("%w, %w, %q", ErrCleanExprNotSafe, ErrCleanSuppRead, expr)

		case 'e':
			return fmt.Errorf("%w, %w, %q", ErrCleanExprNotSafe, ErrCleanSuppExec, expr)

		default:
			return fmt.Errorf("%w, %w, %q, %q", ErrCleanExprNotSafe, ErrCleanSuppCmd, string(tok.cmd), expr)
		}
	}

	return nil
}

// skipDelim discards until unescaped delim.
func skipDelim(rdr *bufio.Reader, delim rune) error {
	var prev rune

	for {
		ch, _, err := rdr.ReadRune()

		switch {
		case err != nil:
			return fmt.Errorf("%w, %c", ErrCleanDelimInvalid, delim)

		case ch == '\n':
			return fmt.Errorf("%w, %c", ErrCleanDelimInvalid, delim)

		case ch == delim && prev != '\\':
			return nil
		}

		prev = ch
	}
}

// parseSub reads delimit, pattern and repl to return flags.
func parseSub(rdr *bufio.Reader) (string, error) {
	delim, _, err := rdr.ReadRune()
	if err != nil {
		return "", ErrCleanDelimInvalid
	}

	if err = skipDelim(rdr, delim); err != nil {
		return "", fmt.Errorf("%w, %v", ErrCleanDelimInvalid, err)
	}

	if err = skipDelim(rdr, delim); err != nil {
		return "", fmt.Errorf("%w, %v", ErrCleanDelimInvalid, err)
	}

	return flagsID(rdr), nil
}

func parseTransl(rdr *bufio.Reader) error {
	delim, _, err := rdr.ReadRune()
	if err != nil {
		return ErrCleanDelimInvalid
	}

	if err = skipDelim(rdr, delim); err != nil {
		return fmt.Errorf("%w, %v", ErrCleanDelimInvalid, err)
	}

	if err = skipDelim(rdr, delim); err != nil {
		return fmt.Errorf("%w, %v", ErrCleanDelimInvalid, err)
	}

	return nil
}

// flagsID skips leading wspace from flags from subs.
// WriteRune always nil https://pkg.go.dev/strings#Builder.WriteRune
func flagsID(rdr *bufio.Reader) string {
	var buf strings.Builder

	for {
		ch, _, err := rdr.ReadRune()
		if err != nil {
			return buf.String()
		}

		if ch == ' ' || ch == '\t' {
			continue
		}

		if ch == ';' || ch == '\n' || ch == '\r' {
			// lint
			_ = rdr.UnreadRune()

			return buf.String()
		}

		buf.WriteRune(ch)

		break
	}
	for {
		ch, _, err := rdr.ReadRune()
		if err != nil {
			break
		}

		if ch == ';' || ch == '\n' || ch == ' ' || ch == '\t' || ch == '\r' {
			// lint
			_ = rdr.UnreadRune()

			break
		}

		buf.WriteRune(ch)
	}

	return buf.String()
}

// skipIdent discards id (label name, filename, etc)
func skipIdent(rdr *bufio.Reader) {
	for {
		ch, _, err := rdr.ReadRune()
		if err != nil {
			return
		}

		if ch == ' ' || ch == '\t' {
			continue
		}

		if ch == ';' || ch == '\n' {
			// lint
			_ = rdr.UnreadRune()

			return
		}

		// first non space char, continue.
		break
	}
	for {
		ch, _, err := rdr.ReadRune()
		if err != nil {
			return
		}

		if ch == ';' || ch == '\n' || ch == ' ' || ch == '\t' {
			// lint
			_ = rdr.UnreadRune()

			return
		}
	}
}

// skipMultiLn discards until unescaped end of line.
func skipMultiLn(rdr *bufio.Reader) {
	for {
		ln, err := rdr.ReadString('\n')
		if err != nil || len(ln) == 0 {
			return
		}

		// doesn't end with \ before newline, we're done.
		if trim := strings.TrimRight(ln, "\n\r"); len(trim) == 0 || trim[len(trim)-1] != '\\' {
			return
		}
	}
}

// isSafeFlags checks for safe sub flags.
func isSafeFlags(flags string, expr string) error {
	for _, char := range flags {
		switch char {
		// safes
		case 'g', 'i', 'p', 'm', 'I':
			continue

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			continue

		// unsafes
		case 'w', 'W':
			return fmt.Errorf("%w, %w, %q", ErrCleanExprNotSafe, ErrCleanSuppWrite, expr)

		case 'e':
			return fmt.Errorf("%w, %w, %q", ErrCleanExprNotSafe, ErrCleanSuppExec, expr)

		default:
			return fmt.Errorf("%w, %w, %q, %q", ErrCleanExprNotSafe, ErrCleanSuppFlag, string(char), expr)
		}
	}

	return nil
}

func lexExpr(expr string) ([]token, error) {
	var (
		rdr  = bufio.NewReader(strings.NewReader(expr))
		toks []token
		err  error
	)

	for {
		var char rune

		char, _, err = rdr.ReadRune()
		if err != nil {
			break
		}

		switch char {
		case ' ', '\t', '\r':
			continue

		case '\n':
			toks = append(
				toks,

				token{
					cmd: ';',
				},
			)

			continue

		case '#':
			for {
				char, _, err := rdr.ReadRune()
				if err != nil || char == '\n' {
					break
				}
			}

			toks = append(
				toks,

				token{
					cmd: ';',
				},
			)

			continue

		case ';':
			toks = append(
				toks,

				token{
					cmd: ';',
				},
			)

		case ',':
			toks = append(
				toks,

				token{
					cmd: ',',
				},
			)

		case '{':
			toks = append(
				toks,

				token{
					cmd: '{',
				},
			)

		case '}':
			toks = append(
				toks,

				token{
					cmd: '}',
				},
			)

		case '!':
			toks = append(
				toks,

				token{
					cmd: '!',
				},
			)

		case '$':
			toks = append(
				toks,

				token{
					cmd: '$',
				},
			)

		case '/':
			if err = skipDelim(rdr, '/'); err != nil {
				return nil, fmt.Errorf("%w, %v", ErrCleanReInvalid, err)
			}

			toks = append(
				toks,

				token{
					cmd: '/',
				},
			)

		case ':':
			skipIdent(rdr)

			toks = append(
				toks,

				token{
					cmd: ':',
				},
			)

		case 'b', 't':
			skipIdent(rdr)

			toks = append(
				toks,

				token{
					cmd: char,
				},
			)

		case 's':
			flags, err := parseSub(rdr)
			if err != nil {
				return nil, fmt.Errorf("%w, %v", ErrCleanSubInvalid, err)
			}

			toks = append(
				toks,

				token{
					cmd:   's',
					flags: flags,
				},
			)

		case 'y':
			if err = parseTransl(rdr); err != nil {
				return nil, fmt.Errorf("%w, %v", ErrCleanExprTranslInvalid, err)
			}

			toks = append(
				toks,

				token{
					cmd: 'y',
				},
			)

		case 'c', 'i', 'a':
			skipMultiLn(rdr)

			toks = append(
				toks,

				token{
					cmd: char,
				},
			)

		case 'r', 'R', 'w', 'W':
			skipIdent(rdr)

			toks = append(
				toks,

				token{
					cmd: char,
				},
			)

		default:
			if char >= '0' && char <= '9' {
				for {
					next, _, err := rdr.ReadRune()
					if err != nil {
						break
					}

					if next < '0' || next > '9' {
						// lint
						_ = rdr.UnreadRune()

						break
					}
				}

				toks = append(
					toks,

					token{
						cmd: char,
					},
				)
			} else {
				toks = append(
					toks,

					token{
						cmd: char,
					},
				)
			}
		}
	}

	if err == io.EOF {
		err = nil
	}

	return toks, err
}
