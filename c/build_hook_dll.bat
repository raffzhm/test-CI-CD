@echo off
echo Compiling MouseHook DLL...
gcc -shared -o MouseHook.dll mouse_hook.c -luser32 -lkernel32 -Wl,--subsystem,windows
echo DLL compilation complete!