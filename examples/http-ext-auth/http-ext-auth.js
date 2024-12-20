// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

const Http = require("http");
const path = require("path");

const tokens = {
    "token1": "user1",
    "token2": "user2",
    "token3": "user3"
};

const server = new Http.Server((req, res) => {
    const authorization = req.headers["authorization"] || "";
    const extracted = authorization.split(" ");
    if (extracted.length === 2 && extracted[0] === "Bearer") {
        const user = checkToken(extracted[1]);
        console.log(`token: "${extracted[1]}" user: "${user}`);
        if (user !== undefined) {
            // The authorization server returns a response with "x-current-user" header for a successful
            // request.
            res.writeHead(200, { "x-current-user": user });
            return res.end();
        }
    }
    res.writeHead(403);
    res.end();
});

const port = process.env.PORT || 9002;
server.listen(port);
console.log(`starting HTTP server on: ${port}`);

function checkToken(token) {
    return tokens[token];
}