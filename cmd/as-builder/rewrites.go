package main

import (
	"fmt"
	"strings"
)

type matchTarget struct {
	pathToMatch string
	target      string
}

type rewrites struct {
	matchTargets []matchTarget
}

func (r *rewrites) String() string {
	str := "func findRewrite(from string) string {\n"

	if r != nil {
		for _, matchTarget := range r.matchTargets {
			strippedSlash := strings.TrimRight(matchTarget.pathToMatch, "/")
			str +=
				fmt.Sprintf("if from == \"%s\" || from == \"%s\" { return \"%s\" }\n",
					matchTarget.pathToMatch, strippedSlash, matchTarget.target)
		}
	}

	str += "return from\n}\n"

	return str
}

func initRewrites(rewriteParts string) (*rewrites, error) {
	parts := strings.Split(rewriteParts, ",")
	var matchTargets []matchTarget

	for _, rewriteTuple := range parts {
		rewriteTupleParts := strings.Split(rewriteTuple, ":")
		if len(rewriteTupleParts) != 2 || rewriteTupleParts[0] == "" || rewriteTupleParts[1] == "" {
			return nil, fmt.Errorf("Exactly 2 non empty rewrite tuple specification required. \"%s\" is invalid", rewriteTuple)
		}
		matchTargets = append(matchTargets, matchTarget{rewriteTupleParts[0], rewriteTupleParts[1]})
	}

	return &rewrites{matchTargets}, nil
}
