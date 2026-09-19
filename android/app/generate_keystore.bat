@echo off
echo Generating debug keystore...
echo CN=Android Debug,O=Android,C=US > temp_input.txt
keytool -genkey -alias androiddebugkey -keyalg RSA -keysize 2048 -validity 10000 -keystore debug.keystore -storepass android -keypass android -dname "CN=Android Debug,O=Android,C=US"
del temp_input.txt
echo Keystore generated successfully!
