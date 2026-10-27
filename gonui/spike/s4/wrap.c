// wrap.c — то, что удобнее написать на C: регистрация класса с указателем на трамплин
// nuiWndProc (//export в main.go), цикл сообщений, разбор командной строки.
#include "win32.h"

#define S4_RES(id) ((LPCWSTR)(UINT_PTR)(WORD)(id))

static const WCHAR cls[] = {'n', 'u', 'i', 'S', '4', 0};

// comctl32 не импортируется статически: статический импорт загрузил бы её при старте —
// раньше, чем s4_actctx включит контекст с Common Controls v6, — и осталась бы v5.
// Загруженная после контекста библиотека — уже v6 (перенаправление DLL контекстом).
static void s4_comctl(void) {
	static const WCHAR name[] = {'c', 'o', 'm', 'c', 't', 'l', '3', '2', '.', 'd', 'l', 'l', 0};
	INITCOMMONCONTROLSEX icc = {sizeof icc, NUI_ICC_STANDARD_CLASSES};
	HMODULE m = LoadLibraryW(name);
	BOOL (*init)(const INITCOMMONCONTROLSEX *) =
		m ? (BOOL (*)(const INITCOMMONCONTROLSEX *))GetProcAddress(m, "InitCommonControlsEx") : 0;
	if (init)
		init(&icc);
}

HWND s4_window(LPCWSTR title, int w, int h) {
	s4_comctl();
	HINSTANCE inst = GetModuleHandleW(0);
	WNDCLASSEXW wc = {0};
	wc.cbSize = sizeof wc;
	wc.lpfnWndProc = nuiWndProc;
	wc.hInstance = inst;
	wc.hCursor = LoadCursorW(0, S4_RES(32512)); /* IDC_ARROW */
	wc.hbrBackground = GetSysColorBrush(NUI_COLOR_BTNFACE);
	wc.lpszClassName = cls;
	if (!RegisterClassExW(&wc))
		return 0;
	return CreateWindowExW(NUI_WS_EX_CONTROLPARENT, cls, title, NUI_WS_OVERLAPPEDWINDOW,
		(int)0x80000000, (int)0x80000000, w, h, 0, 0, inst, 0); /* CW_USEDEFAULT */
}

HWND s4_child(HWND parent, LPCWSTR c, LPCWSTR text, DWORD style, DWORD exstyle, int x, int y, int w, int h) {
	HWND hw = CreateWindowExW(exstyle, c, text, NUI_WS_CHILD | NUI_WS_VISIBLE | style, x, y, w, h, parent, 0,
		GetModuleHandleW(0), 0);
	SendMessageW(hw, NUI_WM_SETFONT, (WPARAM)GetStockObject(NUI_DEFAULT_GUI_FONT), 1);
	return hw;
}

// Цикл как в driver-win32.md, §3: IsDialogMessage даёт Tab, Enter и Esc.
int s4_loop(HWND dlg) {
	MSG m;
	while (GetMessageW(&m, 0, 0, 0) > 0) {
		if (IsDialogMessageW(dlg, &m))
			continue;
		TranslateMessage(&m);
		DispatchMessageW(&m);
	}
	return (int)m.wParam;
}

// s4_selftest: в командной строке есть «-selftest» — без пакета os на стороне Go.
int s4_selftest(void) {
	static const WCHAR key[] = {'-', 's', 'e', 'l', 'f', 't', 'e', 's', 't', 0};
	for (LPCWSTR p = GetCommandLineW(); *p; p++) {
		int i = 0;
		while (key[i] && p[i] == key[i])
			i++;
		if (!key[i])
			return 1;
	}
	return 0;
}

// Манифест (RT_MANIFEST = 24, id 1) попал в exe — значит, .syso подхватился.
int s4_has_manifest(void) { return FindResourceW(0, S4_RES(1), S4_RES(24)) != 0; }

// Запасной путь без манифеста в exe (driver-win32.md, §8): контекст активации из манифеста
// shell32.dll (ресурс 124, Common Controls v6) становится контекстом процесса по умолчанию.
// Если у exe свой манифест, система вернёт ошибку и останется при нём. 0 — получилось.
DWORD s4_actctx(void) {
	static const WCHAR shell32[] = {'s', 'h', 'e', 'l', 'l', '3', '2', '.', 'd', 'l', 'l', 0};
	WCHAR dir[260];
	ACTCTXW act = {0};
	ULONG_PTR cookie;
	if (!GetSystemDirectoryW(dir, 260))
		return GetLastError();
	act.cbSize = sizeof act;
	act.dwFlags = NUI_ACTCTX_FLAG_RESOURCE_NAME_VALID | NUI_ACTCTX_FLAG_SET_PROCESS_DEFAULT |
		NUI_ACTCTX_FLAG_ASSEMBLY_DIRECTORY_VALID;
	act.lpSource = shell32;
	act.lpAssemblyDirectory = dir;
	act.lpResourceName = S4_RES(124);
	HANDLE h = CreateActCtxW(&act);
	if (h == (HANDLE)(LONG_PTR)-1)
		return GetLastError();
	return ActivateActCtx(h, &cookie) ? 0 : GetLastError();
}
