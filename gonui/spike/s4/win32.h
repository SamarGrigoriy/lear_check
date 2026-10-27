/* win32.h — собственные прототипы Win32 для спайка S4 (ADR-3): только то, что нужно hello.
   В режиме NUI_CHECK (check.sh) всё сверяется с <windows.h> из mingw-w64. */
#ifndef S4_WIN32_H
#define S4_WIN32_H

#ifndef NUI_CHECK
typedef __WCHAR_TYPE__ WCHAR;
typedef const WCHAR *LPCWSTR;
typedef WCHAR *LPWSTR;
typedef int BOOL, INT;
typedef unsigned int UINT;
typedef unsigned short WORD, ATOM;
typedef unsigned long DWORD;
typedef long LONG;
typedef long long LONG_PTR;
typedef unsigned long long UINT_PTR, ULONG_PTR;
typedef unsigned long ULONG;
typedef unsigned short USHORT, LANGID;
typedef UINT_PTR WPARAM;
typedef LONG_PTR LPARAM, LRESULT;
typedef void *LPVOID, *HANDLE;
typedef struct HWND__ *HWND;
typedef struct HINSTANCE__ *HINSTANCE;
typedef HINSTANCE HMODULE;
typedef struct HMENU__ *HMENU;
typedef struct HICON__ *HICON;
typedef HICON HCURSOR;
typedef struct HBRUSH__ *HBRUSH;
typedef struct HRSRC__ *HRSRC;
typedef void *HGDIOBJ;
typedef long long (*FARPROC)(void);
typedef LRESULT (*WNDPROC)(HWND, UINT, WPARAM, LPARAM);
typedef void (*TIMERPROC)(HWND, UINT, UINT_PTR, DWORD);
typedef struct tagPOINT { LONG x, y; } POINT;
typedef struct tagMSG {
	HWND hwnd;
	UINT message;
	WPARAM wParam;
	LPARAM lParam;
	DWORD time;
	POINT pt;
} MSG;
typedef struct tagWNDCLASSEXW {
	UINT cbSize, style;
	WNDPROC lpfnWndProc;
	int cbClsExtra, cbWndExtra;
	HINSTANCE hInstance;
	HICON hIcon;
	HCURSOR hCursor;
	HBRUSH hbrBackground;
	LPCWSTR lpszMenuName, lpszClassName;
	HICON hIconSm;
} WNDCLASSEXW;
typedef struct tagINITCOMMONCONTROLSEX { DWORD dwSize, dwICC; } INITCOMMONCONTROLSEX;
typedef struct tagACTCTXW {
	ULONG cbSize;
	DWORD dwFlags;
	LPCWSTR lpSource;
	USHORT wProcessorArchitecture;
	LANGID wLangId;
	LPCWSTR lpAssemblyDirectory, lpResourceName, lpApplicationName;
	HMODULE hModule;
} ACTCTXW;
#endif

/* Раскладка структур: в обоих режимах — одни и те же размеры. */
_Static_assert(sizeof(MSG) == 48, "MSG");
_Static_assert(sizeof(WNDCLASSEXW) == 80, "WNDCLASSEXW");
_Static_assert(sizeof(ACTCTXW) == 56, "ACTCTXW");

#ifdef NUI_CHECK
#define NUI_FN(ret, name, params) \
	_Static_assert(__builtin_types_compatible_p(__typeof__(&name), ret (__stdcall *) params), #name);
#define NUI_CONST(name, val) \
	_Static_assert((val) == (name), #name); enum { NUI_##name = (val) };
#else
#define NUI_FN(ret, name, params) ret name params;
#define NUI_CONST(name, val) enum { NUI_##name = (val) };
#endif

NUI_CONST(WM_DESTROY, 0x0002)
NUI_CONST(WM_CLOSE, 0x0010)
NUI_CONST(WM_SETFONT, 0x0030)
NUI_CONST(WM_COMMAND, 0x0111)
NUI_CONST(WM_TIMER, 0x0113)
NUI_CONST(BM_CLICK, 0x00F5)
NUI_CONST(BN_CLICKED, 0)
NUI_CONST(EM_SETCUEBANNER, 0x1501)
NUI_CONST(WS_OVERLAPPEDWINDOW, 0x00CF0000)
NUI_CONST(WS_CHILD, 0x40000000)
NUI_CONST(WS_VISIBLE, 0x10000000)
NUI_CONST(WS_TABSTOP, 0x00010000)
NUI_CONST(WS_EX_CONTROLPARENT, 0x00010000)
NUI_CONST(WS_EX_CLIENTEDGE, 0x00000200)
NUI_CONST(BS_DEFPUSHBUTTON, 0x00000001)
NUI_CONST(SS_NOPREFIX, 0x00000080)
NUI_CONST(ES_AUTOHSCROLL, 0x0080)
NUI_CONST(SW_SHOWNORMAL, 1)
NUI_CONST(COLOR_BTNFACE, 15)
NUI_CONST(DEFAULT_GUI_FONT, 17)
NUI_CONST(ICC_STANDARD_CLASSES, 0x00004000)
NUI_CONST(ACTCTX_FLAG_ASSEMBLY_DIRECTORY_VALID, 0x004)
NUI_CONST(ACTCTX_FLAG_RESOURCE_NAME_VALID, 0x008)
NUI_CONST(ACTCTX_FLAG_SET_PROCESS_DEFAULT, 0x010)

NUI_FN(HMODULE, GetModuleHandleW, (LPCWSTR lpModuleName))
NUI_FN(LPWSTR, GetCommandLineW, (void))
NUI_FN(HRSRC, FindResourceW, (HMODULE hModule, LPCWSTR lpName, LPCWSTR lpType))
NUI_FN(ATOM, RegisterClassExW, (const WNDCLASSEXW *wc))
NUI_FN(HWND, CreateWindowExW, (DWORD dwExStyle, LPCWSTR lpClassName, LPCWSTR lpWindowName, DWORD dwStyle,
       int X, int Y, int nWidth, int nHeight, HWND hWndParent, HMENU hMenu, HINSTANCE hInstance, LPVOID lpParam))
NUI_FN(LRESULT, DefWindowProcW, (HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam))
NUI_FN(BOOL, ShowWindow, (HWND hWnd, int nCmdShow))
NUI_FN(BOOL, GetMessageW, (MSG *lpMsg, HWND hWnd, UINT wMsgFilterMin, UINT wMsgFilterMax))
NUI_FN(BOOL, TranslateMessage, (const MSG *lpMsg))
NUI_FN(LRESULT, DispatchMessageW, (const MSG *lpMsg))
NUI_FN(BOOL, IsDialogMessageW, (HWND hDlg, MSG *lpMsg))
NUI_FN(void, PostQuitMessage, (int nExitCode))
NUI_FN(BOOL, DestroyWindow, (HWND hWnd))
NUI_FN(LRESULT, SendMessageW, (HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam))
NUI_FN(BOOL, SetWindowTextW, (HWND hWnd, LPCWSTR lpString))
NUI_FN(UINT_PTR, SetTimer, (HWND hWnd, UINT_PTR nIDEvent, UINT uElapse, TIMERPROC lpTimerFunc))
NUI_FN(BOOL, KillTimer, (HWND hWnd, UINT_PTR uIDEvent))
NUI_FN(HCURSOR, LoadCursorW, (HINSTANCE hInstance, LPCWSTR lpCursorName))
NUI_FN(HBRUSH, GetSysColorBrush, (int nIndex))
NUI_FN(HGDIOBJ, GetStockObject, (int i))
NUI_FN(BOOL, InitCommonControlsEx, (const INITCOMMONCONTROLSEX *picce)) /* только через GetProcAddress */
NUI_FN(HMODULE, LoadLibraryW, (LPCWSTR lpLibFileName))
NUI_FN(HANDLE, CreateActCtxW, (const ACTCTXW *pActCtx))
NUI_FN(BOOL, ActivateActCtx, (HANDLE hActCtx, ULONG_PTR *lpCookie))
NUI_FN(UINT, GetSystemDirectoryW, (LPWSTR lpBuffer, UINT uSize))
NUI_FN(DWORD, GetLastError, (void))
NUI_FN(FARPROC, GetProcAddress, (HMODULE hModule, const char *lpProcName))

/* wrap.c */
HWND   s4_window(LPCWSTR title, int w, int h);
HWND   s4_child(HWND parent, LPCWSTR cls, LPCWSTR text, DWORD style, DWORD exstyle, int x, int y, int w, int h);
int    s4_loop(HWND dlg);
int    s4_selftest(void);
int    s4_has_manifest(void);
DWORD  s4_actctx(void);
LRESULT nuiWndProc(HWND hwnd, UINT msg, WPARAM wp, LPARAM lp); /* //export в main.go */

#endif
