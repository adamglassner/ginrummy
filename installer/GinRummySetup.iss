; Gin Rummy - standalone setup. Build: bash build.sh, then ISCC.exe /DGameVersion=<version> installer\GinRummySetup.iss
; (c) 2026 Ironite
#define GameName    "Gin Rummy"
#define Bundle      "GinRummy"
#define Exe         "GinRummy.exe"
#ifndef GameVersion
  #define GameVersion "1.0.0"
#endif
#define AppGuid     "{{9B2E6F14-3C7A-4D58-A1E9-6F0B2C8D4E73}"
#define Art         "ginrummy"
#define Tagline     "A club car card game. Play a passenger aboard the Starlight Limited, share one PC, or play a friend online."
#define SaveWhat    "Gin Rummy record and settings"
#include "GameSetup.iss"

[Run]
; Let Gin Rummy host online games without a firewall prompt.
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Gin Rummy"" dir=in action=allow program=""{app}\Games\{#Bundle}\{#Exe}"" enable=yes"; Flags: runhidden; StatusMsg: "Letting Gin Rummy host online games..."

[UninstallRun]
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Gin Rummy"""; Flags: runhidden; RunOnceId: "GinRummyFirewall"
