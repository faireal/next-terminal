// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package ldap

import (
	"github.com/ergoapi/ldap"
	"github.com/spf13/viper"
	"next-terminal/models"
	"strings"
)

func LdapLogin(user, pass string) (*models.User, error) {
	ldapconf := ldap.LdapConf{
		LdapURL:               viper.GetString("ldap.ldap_url"),
		LdapSearchDn:          viper.GetString("ldap.ldap_search_dn"),
		LdapSearchPassword:    viper.GetString("ldap.ldap_search_password"),
		LdapBaseDn:            viper.GetString("ldap.ldap_base_dn"),
		LdapFilter:            viper.GetString("ldap.ldap_filter"),
		LdapUID:               viper.GetString("ldap.ldap_uid"),
		LdapScope:             viper.GetInt("ldap.ldap_scope"),
		LdapConnectionTimeout: viper.GetInt("ldap.ldap_connection_timeout"),
	}
	sr, err := ldapconf.LdapReq(user, pass)
	if err != nil {
		return nil, err
	}
	var u models.User
	for _, attr := range sr.Entries[0].Attributes {
		key := strings.ToLower(attr.Name)
		val := attr.Values[0]
		switch key {
		case "displayname":
			u.Nickname = val
		case "mail":
			u.Mail = val
		case "email":
			u.Mail = val
		case "department":
			u.Department = val
		}
	}
	u.Username = user
	return &u, nil
}
