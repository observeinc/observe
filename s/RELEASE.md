Release process:

1. make sure you have tagged the release and pushed to github:
   git tag v0.1.2-r3 HEAD
   git push origin HEAD:main
   git push --tags

2. you must do this on a MacBook, because windows/linux can't notarize software for Macs for Apple reasons.

3. source the appropriate secrets into environment for AC_TEAMID, GITHUB_TOKEN, AUTHENTICODE_KEY etc
   source ~/secrets.env

4. run goreleaser in the root of the repository to build but not publish
   goreleaser

5. Assuming you have the Windows code signing certificates, the Apple developer
   account app specific password, and the right cert/key for notarization all
   available in the environment, this will build, sign, and upload the release
   and binaries to GitHub.

6. Figure out how to homebrew / chocolatey / ...
