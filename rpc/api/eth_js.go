// Copyright 2015 The go-coupe Authors
// This file is part of the go-coupe library.
//
// The go-coupe library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-coupe library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-coupe library. If not, see <http://www.gnu.org/licenses/>.

package api

// JS api provided by web3.js
// nch_sign not standard

const nch_JS = `
web3._extend({
	property: 'eth',
	methods:
	[
		new web3._extend.Method({
			name: 'sign',
			call: 'nch_sign',
			params: 2,
			inputFormatter: [web3._extend.utils.toAddress, null]
		}),
		new web3._extend.Method({
			name: 'resend',
			call: 'nch_resend',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Method({
			name: 'getNatSpec',
			call: 'nch_getNatSpec',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'signTransaction',
			call: 'nch_signTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'submitTransaction',
			call: 'nch_submitTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		})
	],
	properties:
	[
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'nch_pendingTransactions'
		})
	]
});
`
