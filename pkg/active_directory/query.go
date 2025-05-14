/*
# Active Directory

This package initializes all the methods for queries to filter {Active Directory/LDAP}:
- https://learn.microsoft.com/en-us/windows/win32/adsi/search-filter-syntax
- https://theitbros.com/ldap-query-examples-active-directory/

:Copyright: (c) 2024 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/active_directory/query.go
package active_directory

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

const (
	LDAP_MATCHING_RULE_BIT_AND      = "1.2.840.113556.1.4.803"  // A match is found only if all bits from the attribute match the value. This rule is equivalent to a bitwise AND operator.
	LDAP_MATCHING_RULE_BIT_OR       = "1.2.840.113556.1.4.804"  // A match is found if any bits from the attribute match the value. This rule is equivalent to a bitwise OR operator.
	LDAP_MATCHING_RULE_IN_CHAIN     = "1.2.840.113556.1.4.1941" // This rule is limited to filters that apply to the DN. This is a special "extended" match operator that walks the chain of ancestry in objects all the way to the root until it finds a match.
	LDAP_MATCHING_RULE_DN_WITH_DATA = "1.2.840.113556.1.4.2253" // https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/e5bfc285-05b9-494e-a123-c5c4341c450e

	FILTER_USER_ACTIVE                 = "(&(objectCategory=person)(objectClass=user)(!userAccountControl:" + LDAP_MATCHING_RULE_BIT_AND + ":=2))"    // List all active users
	FILTER_USER_ADMIN                  = "(&(objectCategory=person)(objectClass=user)(adminCount=1))"                                                 // List all users in privileged groups [Domain Admins, Enterprise Admins, etc.]
	FILTER_USER_DISABLED               = "(&(objectCategory=person)(objectClass=user)(useraccountcontrol:" + LDAP_MATCHING_RULE_BIT_AND + ":=2))"     // List all disabled users
	FILTER_USER_NESTED_GROUP           = "(&(objectClass=user)(memberOf:" + LDAP_MATCHING_RULE_IN_CHAIN + ":=CN=%s,%s,%s))"                           // To Do: Implement better DN handling
	FILTER_USER_LOCKED                 = "(&(objectCategory=person)(objectClass=user)(lockoutTime>=1))"                                               // List all locked users
	FILTER_USER_PASSWORD_NEVER_EXPIRES = "(&(objectCategory=person)(objectClass=user)(useraccountcontrol:" + LDAP_MATCHING_RULE_BIT_AND + ":=65536))" // List all users with passwords that never expire
)

// LDAPQuery holds parameters for LDAP searches
type LDAPQuery struct {
	BaseDN                 string         // Base Distinguished Name for the search
	Filter                 string         // LDAP search filter
	Attributes             []string       // Attributes to be retrieved
	Scope                  int            // Scope of the search (Base, SingleLevel, WholeSubtree)
	DerefAliases           int            // Behavior regarding alias dereferencing
	SizeLimit              int            // Maximum number of entries to be returned
	TimeLimit              int            // Time limit (in seconds) for the search
	TypesOnly              bool           // Return attribute types only, not values
	Controls               []ldap.Control // Request controls for additional features/behaviors
	PagingSize             uint32         // Size of the paging to be used if any
	validateFilterNotEmpty bool
}

type Filter struct {
	ObjectCategory string
	ObjectClass    string
	MemberOf       string
}

func ConvertAttributes(attributes *[]Attribute) []string {
	var strings []string
	for _, attr := range *attributes {
		strings = append(strings, string(attr))
	}
	return strings
}

// NewLDAPQuery creates a default LDAPQuery with common settings
func NewLDAPQuery(baseDN, filter string, attributes []string) *LDAPQuery {
	return &LDAPQuery{
		BaseDN:       baseDN,
		Filter:       filter,
		Attributes:   attributes,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0, // No limit
		TimeLimit:    0, // No limit
		TypesOnly:    false,
		Controls:     nil,
		PagingSize:   1000, // Common default value for paging
	}
}

// SetBaseDN sets the base DN for the LDAP query
func (q *LDAPQuery) SetBaseDN(baseDN string) *LDAPQuery {
	q.BaseDN = baseDN
	return q
}

// SetFilter sets the search filter for the LDAP query
func (q *LDAPQuery) SetFilter(filter string) *LDAPQuery {
	q.Filter = filter
	return q
}

// SetAttributes sets the attributes to retrieve
func (q *LDAPQuery) SetAttributes(attrs []string) *LDAPQuery {
	q.Attributes = attrs
	return q
}

// SetScope sets the search scope
func (q *LDAPQuery) SetScope(scope int) *LDAPQuery {
	q.Scope = scope
	return q
}

// SetDerefAliases sets the behavior for alias dereferencing
func (q *LDAPQuery) SetDerefAliases(deref int) *LDAPQuery {
	q.DerefAliases = deref
	return q
}

// SetSizeLimit sets the maximum number of entries to return
func (q *LDAPQuery) SetSizeLimit(limit int) *LDAPQuery {
	q.SizeLimit = limit
	return q
}

// SetTimeLimit sets the time limit for the search
func (q *LDAPQuery) SetTimeLimit(limit int) *LDAPQuery {
	q.TimeLimit = limit
	return q
}

// SetTypesOnly specifies if only attribute types should be returned
func (q *LDAPQuery) SetTypesOnly(typesOnly bool) *LDAPQuery {
	q.TypesOnly = typesOnly
	return q
}

// SetPagingSize sets the size of the paging
func (q *LDAPQuery) SetPagingSize(size uint32) *LDAPQuery {
	q.PagingSize = size
	return q
}

// Validate checks if the LDAP query parameters are set correctly
func (q *LDAPQuery) Validate() error {
	if q.validateFilterNotEmpty && q.Filter == "" {
		return errors.New("LDAP query filter cannot be empty")
	}

	if q.Scope != ldap.ScopeBaseObject &&
		q.Scope != ldap.ScopeSingleLevel &&
		q.Scope != ldap.ScopeWholeSubtree {
		return fmt.Errorf("invalid LDAP query scope: %d", q.Scope)
	}

	if q.DerefAliases != ldap.NeverDerefAliases &&
		q.DerefAliases != ldap.DerefInSearching &&
		q.DerefAliases != ldap.DerefFindingBaseObj &&
		q.DerefAliases != ldap.DerefAlways {
		return fmt.Errorf("invalid LDAP deref aliases setting: %d", q.DerefAliases)
	}

	if q.SizeLimit < 0 {
		return errors.New("LDAP query size limit cannot be negative")
	}

	if q.TimeLimit < 0 {
		return errors.New("LDAP query time limit cannot be negative")
	}

	if q.PagingSize < 0 {
		return errors.New("LDAP query paging size cannot be negative")
	}

	// Additional validations can be added here as needed

	return nil
}

// IsEmpty checks if the query parameters are empty
func (q *LDAPQuery) IsEmpty() bool {
	return q.BaseDN == "" && q.Filter == ""
}
