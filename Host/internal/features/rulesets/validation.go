package rulesets

import domainrulesets "github.com/jvanspijk/SocialGamesHoster/Host/internal/domain/rulesets"

type (
	ValidationIssue  = domainrulesets.ValidationIssue
	ValidationReport = domainrulesets.ValidationReport
)

var stableIDPattern = domainrulesets.StableIDPattern

func Validate(def DefinitionV1, assetKeys map[string]struct{}) ValidationReport {
	return domainrulesets.Validate(def, assetKeys)
}

func MatchingRoles(roles []Role, selector Selector) []Role {
	return domainrulesets.MatchingRoles(roles, selector)
}

func selectorMatches(role Role, selector Selector) bool {
	return domainrulesets.SelectorMatches(role, selector)
}
