@echo off

set project_root=%~dp0%
set go_path="C:\Program Files\Go\bin"
:: NOTE(nick): make sure gcc.exe is in your path (for CGO build)
set gcc_path="C:\ProgramData\chocolatey\bin\"
set PATH=%PATH%;%gcc_path%

:: NOTE(nick): these need to be in the same directory as the final executable
copy /y %project_root%\lib\win32\automerge.dll %project_root%
copy /y %project_root%\lib\win32\automerge.lib %project_root%

set CGO_LDFLAGS=%project_root%\lib\win32\automerge.dll

pushd %project_root%
  %go_path%\go.exe run .
popd

exit /B %errorlevel%
