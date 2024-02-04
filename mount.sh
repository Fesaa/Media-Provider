echo username=$USER >> /app/.smbcredentials
echo password=$PASS >> /app/.smbcredentials
echo domain=$DOMAIN >> /app/.smbcredentials

mount -t cifs -o credentials=/app/.smbcredentials //$URL /app/mount