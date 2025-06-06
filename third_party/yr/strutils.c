/*
Copyright (c) 2007-2014. The YARA Authors. All Rights Reserved.

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

#include <stdio.h>
#include <string.h>
#include <yr_mem.h>
#include <yr_strutils.h>

uint64_t xtoi(const char* hexstr)
{
  size_t i;
  size_t l = strlen(hexstr);

  uint64_t r = 0;

  for (i = 0; i < l; i++)
  {
    switch (hexstr[i])
    {
    case '0':
    case '1':
    case '2':
    case '3':
    case '4':
    case '5':
    case '6':
    case '7':
    case '8':
    case '9':
      r |= ((uint64_t) (hexstr[i] - '0')) << ((l - i - 1) * 4);
      break;
    case 'a':
    case 'b':
    case 'c':
    case 'd':
    case 'e':
    case 'f':
      r |= ((uint64_t) (hexstr[i] - 'a' + 10)) << ((l - i - 1) * 4);
      break;
    case 'A':
    case 'B':
    case 'C':
    case 'D':
    case 'E':
    case 'F':
      r |= ((uint64_t) (hexstr[i] - 'A' + 10)) << ((l - i - 1) * 4);
      break;
    default:
      i = l;  // force loop exit
    }
  }

  return r;
}

/*

strlcpy and strlcat are defined in FreeBSD and OpenBSD,
the following implementations were taken from OpenBSD.

*/

#if !HAVE_STRLCPY && !defined(strlcpy)
size_t strlcpy(char* dst, const char* src, size_t size)
{
  register char* d = dst;
  register const char* s = src;
  register size_t n = size;

  // Copy as many bytes as will fit

  if (n != 0 && --n != 0)
  {
    do
    {
      if ((*d++ = *s++) == 0)
        break;

    } while (--n != 0);
  }

  // Not enough room in dst, add NUL and traverse rest of src

  if (n == 0)
  {
    if (size != 0)
      *d = '\0';  // NULL-terminate dst

    while (*s++)
      ;
  }

  return (s - src - 1);  // count does not include NULL
}
#endif

#if !HAVE_STRLCAT && !defined(strlcat)
size_t strlcat(char* dst, const char* src, size_t size)
{
  register char* d = dst;
  register const char* s = src;
  register size_t n = size;
  size_t dlen;

  // Find the end of dst and adjust bytes left but don't go past end

  while (n-- != 0 && *d != '\0') d++;

  dlen = d - dst;
  n = size - dlen;

  if (n == 0)
    return (dlen + strlen(s));

  while (*s != '\0')
  {
    if (n != 1)
    {
      *d++ = *s;
      n--;
    }
    s++;
  }

  *d = '\0';

  return (dlen + (s - src));  // count does not include NULL
}
#endif

int strnlen_w(const char* w_str)
{
  int len = 0;

  while (w_str[0] || w_str[1])
  {
    w_str += 2;
    len += 1;
  }

  return len;
}

int strcmp_w(const char* w_str, const char* str)
{
  while (*str != 0 && w_str[0] == *str && w_str[1] == 0)
  {
    w_str += 2;
    str += 1;
  }

  // Higher-order byte of wide char non-zero? -> w_str is larger than str

  if (w_str[1] != 0)
    return 1;

  return w_str[0] - *str;
}

size_t strlcpy_w(char* dst, const char* w_src, size_t n)
{
  register char* d = dst;
  register const char* s = w_src;

  while (n > 1 && *s != 0)
  {
    *d = *s;
    d += 1;
    n -= 1;
    s += 2;
  }

  while (*s) s += 2;

  *d = '\0';

  return (s - w_src) / 2;
}

#if !HAVE_MEMMEM && !defined(memmem)
void* memmem(
    const void* haystack,
    size_t haystack_size,
    const void* needle,
    size_t needle_size)
{
  char* sp = (char*) haystack;
  char* pp = (char*) needle;
  char* eos;

  if (haystack == NULL || haystack_size == 0 || needle == NULL ||
      needle_size == 0)
    return NULL;

  eos = sp + haystack_size - needle_size;

  while (sp <= eos)
  {
    if (*sp == *pp && memcmp(sp, pp, needle_size) == 0)
      return sp;

    sp++;
  }

  return NULL;
}
#endif

///////////////////////////////////////////////////////////////////////////////
// This our own implementation of isalnum(). The library version is locale
// dependent in some platforms and can consider non-ASCII characters to be
// alphanumeric.
//
int yr_isalnum(const uint8_t* s)
{
  return (*s >= 0x30 && *s <= 0x39) || (*s >= 0x41 && *s <= 0x5a) ||
         (*s >= 0x61 && *s <= 0x7a);
}

//////////////////////////////////////////////////////////////////////////
// This our own implementation of vasprintf(), as it is not available on
// some platforms. It is based on the implementation of vsnprintf but it
// allocates memory using yr_malloc, and therefore the caller must free
// the memory using yr_free.
//
void yr_vasprintf(char** strp, const char* fmt, va_list ap)
{
  va_list ap_copy;
  va_copy(ap_copy, ap);
  *strp = NULL;

  int len = vsnprintf(NULL, 0, fmt, ap_copy);

  va_end(ap_copy);

  if (len < 0)
    return;

  *strp = (char*) yr_malloc(len + 1);

  if (*strp == NULL)
    return;

  vsnprintf(*strp, len + 1, fmt, ap);
}

//////////////////////////////////////////////////////////////////////////
// This our own implementation of asprintf(), see yr_vasprintf() for details.
//
void yr_asprintf(char** strp, const char* fmt, ...)
{
  va_list ap;
  va_start(ap, fmt);
  yr_vasprintf(strp, fmt, ap);
  va_end(ap);
}
