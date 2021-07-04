#!/bin/bash

# Default Variable Declarations
kPath="./keys/"
ovpnName='client'
NAME='client'

DEFAULT="inline_client.conf"
OVPN_FILE="${kPath}${ovpnName}.ovpn"
CLIENT_CRT="${kPath}${ovpnName}.crt"
CLIENT_KEY="${kPath}${ovpnName}.key"
SERVER_CA="${kPath}ca.crt"
SERVER_TLS_AUTH="../ta.key"

#1st Verify that client's Public Key Exists
if [ ! -f $CLIENT_CRT ]; then
   echo "[ERROR]: Client Public Key Certificate not found: $CLIENT_CRT"
   exit
fi

#Then, verify that there is a private key for that client
if [ ! -f $CLIENT_KEY ]; then
   echo "[ERROR]: Client 3des Private Key not found: $CLIENT_KEY"
   exit
fi

#Confirm the CA public key exists
if [ ! -f $SERVER_CA ]; then
   echo "[ERROR]: CA Public Key not found: $SERVER_CA"
   exit
fi

#Confirm the tls-auth ta key file exists
if [ ! -f $SERVER_TA ]; then
   echo "[ERROR]: tls-auth Key not found: $SERVER_TA"
   exit
fi
echo "tls-auth Private Key found: $SERVER_TA"

#Ready to make a new .opvn file - Start by populating with the

cat <<EOF > $OVPN_FILE
client
dev tun
proto udp
remote 127.0.0.1 1194
resolv-retry infinite
nobind
persist-key
persist-tun
mute-replay-warnings
remote-cert-tls server
cipher AES-256-CBC
comp-lzo
verb 3
;mute 20

ca [inline]
cert [inline]
key [inline]
;tls-auth [inline] 1
EOF

#Now, append the CA Public Cert
echo "<ca>" >> $OVPN_FILE
cat $SERVER_CA | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' >> $OVPN_FILE
echo "</ca>" >> $OVPN_FILE

#Next append the client Public Cert
echo "<cert>" >> $OVPN_FILE
cat $CLIENT_CRT | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' >> $OVPN_FILE
echo "</cert>" >> $OVPN_FILE

#Then, append the client Private Key
echo "<key>" >> $OVPN_FILE
cat $CLIENT_KEY >> $OVPN_FILE
echo "</key>" >> $OVPN_FILE

#Finally, append the TA Private Key
echo "<tls-auth>" >> $OVPN_FILE
cat $SERVER_TA >> $OVPN_FILE
echo "</tls-auth>" >> $OVPN_FILE