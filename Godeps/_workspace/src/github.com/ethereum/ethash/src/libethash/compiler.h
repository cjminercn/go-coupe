/*
  This file is part of cpp-cjminercn.

  cpp-cjminercn is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  cpp-cjminercn is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with cpp-cjminercn.  If not, see <http://www.gnu.org/licenses/>.
*/
/** @file compiler.h
 * @date 2014
 */
#pragma once

// Visual Studio doesn't support the inline keyword in C mode
#if defined(_MSC_VER) && !defined(__cplusplus)
#define inline __inline
#endif

// pretend restrict is a standard keyword
#if defined(_MSC_VER)
#define restrict __restrict
#else
#define restrict __restrict__
#endif

