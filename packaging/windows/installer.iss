#ifndef AppVersion
  #define AppVersion "0.2.5-dev"
#endif
#ifndef WindowsVersion
  #define WindowsVersion "0.2.5.0"
#endif

#define AppName "Social Games Hoster"
#define AppPublisher "Social Games Hoster contributors"
#define AppExeName "SocialGamesHoster.exe"
#define FirewallRuleName "Social Games Hoster"

[Setup]
AppId={{E109D297-C6EF-4CB7-8605-8E5ED1F9627B}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL=https://github.com/jvanspijk/SocialGamesHoster
AppSupportURL=https://github.com/jvanspijk/SocialGamesHoster/issues
AppUpdatesURL=https://github.com/jvanspijk/SocialGamesHoster/releases
DefaultDirName={autopf}\Social Games Hoster
DefaultGroupName=Social Games Hoster
DisableProgramGroupPage=yes
LicenseFile=..\..\LICENSE
OutputDir=..\..\dist
OutputBaseFilename=SocialGamesHoster-{#AppVersion}-windows-x64-setup
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
CloseApplications=yes
RestartApplications=no
AppMutex=Local\SocialGamesHoster
UninstallDisplayIcon={app}\{#AppExeName}
VersionInfoProductName={#AppName}
VersionInfoProductVersion={#WindowsVersion}
VersionInfoDescription=Local-first social game host

[Files]
Source: "..\..\dist\{#AppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\docs\USER_GUIDE.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\..\docs\TROUBLESHOOTING.md"; DestDir: "{app}\docs"; Flags: ignoreversion

[Icons]
Name: "{group}\Social Games Hoster"; Filename: "{app}\{#AppExeName}"
Name: "{group}\Social Games Hoster (Diagnostic Mode)"; Filename: "{app}\{#AppExeName}"; Parameters: "--diagnostics"
Name: "{group}\User Guide"; Filename: "{app}\docs\USER_GUIDE.md"
Name: "{group}\Uninstall Social Games Hoster"; Filename: "{uninstallexe}"

[Run]
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""{#FirewallRuleName}"""; Flags: runhidden; StatusMsg: "Refreshing the private-network firewall rule..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""{#FirewallRuleName}"" dir=in action=allow program=""{app}\{#AppExeName}"" protocol=TCP localport=8090 profile=private enable=yes"; Flags: runhidden; StatusMsg: "Allowing hosting on private networks only..."
Filename: "{app}\{#AppExeName}"; Description: "Launch Social Games Hoster"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""{#FirewallRuleName}"""; Flags: runhidden; RunOnceId: "RemovePrivateFirewallRule"

[Code]
var
  DeleteDataCheckBox: TNewCheckBox;

function InitializeUninstall(): Boolean;
begin
  Result := True;
  DeleteDataCheckBox := TNewCheckBox.Create(UninstallProgressForm);
  DeleteDataCheckBox.Parent := UninstallProgressForm;
  DeleteDataCheckBox.Left := UninstallProgressForm.StatusLabel.Left;
  DeleteDataCheckBox.Top := UninstallProgressForm.StatusLabel.Top + ScaleY(42);
  DeleteDataCheckBox.Width := UninstallProgressForm.StatusLabel.Width;
  DeleteDataCheckBox.Height := ScaleY(36);
  DeleteDataCheckBox.Caption := 'Also permanently delete all games, profiles, rulesets, chat history, and backups';
  DeleteDataCheckBox.Checked := False;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if (CurUninstallStep = usUninstall) and DeleteDataCheckBox.Checked then
  begin
    if MsgBox(
      'Delete all Social Games Hoster data permanently? This cannot be undone. Leave this unchecked to preserve data for a future reinstall.',
      mbConfirmation,
      MB_YESNO
    ) <> IDYES then
      DeleteDataCheckBox.Checked := False;
  end;

  if (CurUninstallStep = usPostUninstall) and DeleteDataCheckBox.Checked then
    DelTree(ExpandConstant('{localappdata}\SocialGamesHoster'), True, True, True);
end;
