#!/usr/bin/env bash

#
# Script to generate some SSL certs, self signed with garbage data.
#  It is suggested you used *actual* SSL certs ..
#
#

export SSL_CERT=ssl.cert
export SSL_KEY=ssl.key
export SSL_PEM=ssl.pem
export SSL_CSR=ssl.csr

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Create ssl certs
#
# https://raymii.org/s/snippets/OpenSSL_generate_CSR_non-interactivemd.html
#
#    /C=NL: 2 letter ISO country code (Netherlands)
#    /ST=: State, Zuid Holland (South holland)
#    /L=: Location, city (Rotterdam)
#    /O=: Organization (Sparkling Network)
#    /OU=: Organizational Unit, Department (IT Department, Sales)
#    /CN=: Common Name, for a website certificate this is the FQDN. (ssl.raymii.org)
#
openssl req -nodes -newkey rsa:2048 -keyout ${DIR}/${SSL_KEY} -out ${DIR}/${SSL_CSR} -subj "/C=XX/ST=X/L=X/O=X/OU=X/CN=*"
openssl x509 -req -days 366 -in  ${DIR}/${SSL_CSR} -signkey ${DIR}/${SSL_KEY} -out  ${DIR}/${SSL_CERT}
cat ${DIR}/${SSL_CERT} ${DIR}/${SSL_KEY} >  ${DIR}/${SSL_PEM}