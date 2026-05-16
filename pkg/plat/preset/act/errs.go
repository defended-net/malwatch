// © Roscoe Skeens <rskeens@defended.net>
// SPDX-License-Identifier: AGPL-3.0-or-later

package act

import "errors"

var (
	// ErrCfgLoad means cfg load error.
	ErrCfgLoad = errors.New("act: cfg load error")

	// ErrOpen means open error.
	ErrOpen = errors.New("act: open error")

	// ErrMv means move err.
	ErrMv = errors.New("act: move error")

	// ErrDel means del err.
	ErrDel = errors.New("act: del error")

	// ErrPathInvalid means invalid path.
	ErrPathInvalid = errors.New("act: invalid path")

	// ErrAttrInvalid means invalid attr.
	ErrAttrInvalid = errors.New("act: invalid attr")
)

// ALERTS
var (
	// ErrAlerterLoad means alerter load err.
	ErrAlerterLoad = errors.New("act: alerter load error")

	// ErrAlertSend means alert send err.
	ErrAlertSend = errors.New("act: alert send error")
)

// QUARANTINES
var (
	// ErrQuarantineNoDir means quarantine dir missing.
	ErrQuarantineNoDir = errors.New("act: no quarantine dir configured in cfg/actions.toml")

	// ErrQuarantineMv means quarantine move err.
	ErrQuarantineMv = errors.New("act: quarantine move err")
)

// CLEANS
var (
	// ErrExprDo means expr do err.
	ErrCleanExprDo = errors.New("act: clean expr do error")

	// ErrCleanExprMissing means missing expr for match.
	ErrCleanExprMissing = errors.New("act: no clean expressions for match")

	// ErrCleanExprMissingToks means missing expr toks.
	ErrCleanExprMissingToks = errors.New("act: no expressions tokens")

	// ErrExprCompile means expr compile err.
	ErrCleanExprCompile = errors.New("act: clean expr compile error")

	// ErrCleanExprLex means expr lex err.
	ErrCleanExprLex = errors.New("act: clean expr lex error")

	// ErrCleanReInvalid means invalid re.
	ErrCleanReInvalid = errors.New("act: invalid regex")

	// ErrCleanDelimInvalid means invalid delim.
	ErrCleanDelimInvalid = errors.New("act: invalid delimit")

	// ErrCleanSubInvalid means invalid sub.
	ErrCleanSubInvalid = errors.New("act: invalid substitution")

	// ErrCleanExprTranslInvalid means invalid translate.
	ErrCleanExprTranslInvalid = errors.New("act: invalid translate")

	// ErrCleanExprNotSafe means unsafe expr.
	ErrCleanExprNotSafe = errors.New("act: unsafe sed expression, only substitution allowed")

	// ErrCleanSubNotSafe means unsafe subs.
	ErrCleanSubNotSafe = errors.New("act: unsafe sed substitution")

	// ErrCleanSuppExec means unsupported exec.
	ErrCleanSuppExec = errors.New("act: exec not supported")

	// ErrCleanSuppRead means unsupported read.
	ErrCleanSuppRead = errors.New("act: read not supported")

	// ErrCleanSuppWrite means unsupported write.
	ErrCleanSuppWrite = errors.New("act: write not supported")

	// ErrCleanSuppCmd means unsupported cmd.
	ErrCleanSuppCmd = errors.New("act: cmd not supported")

	// ErrCleanSuppFlag means unsupported flag.
	ErrCleanSuppFlag = errors.New("act: flag not supported")

	// ErrCleanFail means clean err.
	ErrCleanFail = errors.New("act: clean failed")
)

// EXILES
var (
	// ErrExileNoRegion means exile region missing.
	ErrExileNoRegion = errors.New("act: no region configured for exile")

	// ErrExileUpload means exile upload err.
	ErrExileUpload = errors.New("act: exile upload error")

	// ErrExileDelErr means exile del err.
	ErrExileDelErr = errors.New("act: exile del error")
)
