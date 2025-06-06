/*
Copyright (c) 2007. Victor M. Alvarez [plusvic@gmail.com].

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation and/or
other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

#include <yr_compiler.h>

#undef yyparse
#undef yylex
#undef yyerror
#undef yyfatal
#undef yychar
#undef yydebug
#undef yynerrs
#undef yyget_extra
#undef yyget_lineno

#undef YY_DECL
#undef YY_FATAL_ERROR
#undef YY_EXTRA_TYPE

#define yyparse      yara_yyparse
#define yylex        yara_yylex
#define yyerror      yara_yyerror
#define yyfatal      yara_yyfatal
#define yywarning    yara_yywarning
#define yychar       yara_yychar
#define yydebug      yara_yydebug
#define yynerrs      yara_yynerrs
#define yyget_extra  yara_yyget_extra
#define yyget_lineno yara_yyget_lineno

// Define the ECHO macro as an empty macro in order to avoid the default
// implementation from being used. The default implementation of ECHO
// prints to the console any byte that is not matched by the lexer. It's
// not safe to print random bytes to the console as it may cause the calling
// program to terminate. See: https://github.com/VirusTotal/yara/issues/2007
#define ECHO

#ifndef YY_TYPEDEF_YY_SCANNER_T
#define YY_TYPEDEF_YY_SCANNER_T
typedef void* yyscan_t;
#endif

union YYSTYPE;

#define YY_DECL \
  int yylex(    \
      union YYSTYPE* yylval_param, yyscan_t yyscanner, YR_COMPILER* compiler)

// The default behavior when a fatal error occurs in the parser is calling
// exit(YY_EXIT_FAILURE) for terminating the process. This is not acceptable
// for a library, which should return gracefully to the calling program. For
// this reason we redefine the YY_FATAL_ERROR macro so that it expands to our
// own function instead of the one provided by default.
#define YY_FATAL_ERROR(msg) yara_yyfatal(yyscanner, msg)

#define YY_EXTRA_TYPE YR_COMPILER*
#define YY_USE_CONST

int yyget_lineno(yyscan_t yyscanner);

int yylex(
    union YYSTYPE* yylval_param,
    yyscan_t yyscanner,
    YR_COMPILER* compiler);

int yyparse(void* yyscanner, YR_COMPILER* compiler);

void yyerror(
    yyscan_t yyscanner,
    YR_COMPILER* compiler,
    const char* error_message);

void yywarning(yyscan_t yyscanner, const char* message_fmt, ...)
    YR_PRINTF_LIKE(2, 3);

void yyfatal(yyscan_t yyscanner, const char* error_message);

YY_EXTRA_TYPE yyget_extra(yyscan_t yyscanner);

int yr_lex_parse_rules_bytes(
    const void* rules_data,
    size_t rules_size,
    YR_COMPILER* compiler);

int yr_lex_parse_rules_string(const char* rules_string, YR_COMPILER* compiler);

int yr_lex_parse_rules_file(FILE* rules_file, YR_COMPILER* compiler);

int yr_lex_parse_rules_fd(YR_FILE_DESCRIPTOR rules_fd, YR_COMPILER* compiler);
