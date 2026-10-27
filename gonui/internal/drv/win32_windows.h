/* win32_windows.h — собственные прототипы Win32 (ADR-3, docs/driver-win32.md, §1).
   Только объявления: файл подключается в преамбуле файла с //export.
   В режиме NUI_CHECK (tools/check-protos.sh) всё сверяется с <windows.h> из mingw-w64:
   типы и структуры тогда берутся из настоящих заголовков, а размеры наших структур
   проверяются в обоих режимах. */
#ifndef NUI_WIN32_H
#define NUI_WIN32_H

/* ---------- 1. Базовые типы и структуры (в режиме проверки — из настоящих заголовков) ---------- */
#ifndef NUI_CHECK
typedef __UINTPTR_TYPE__ uintptr_t;
typedef __WCHAR_TYPE__ WCHAR;
typedef WCHAR *LPWSTR;
typedef const WCHAR *LPCWSTR, *PCWSTR;
typedef int BOOL, INT;
typedef unsigned int UINT;
typedef unsigned char BYTE;
typedef unsigned short WORD, ATOM, USHORT, LANGID;
typedef short SHORT;
typedef unsigned long DWORD, ULONG, COLORREF;
typedef long LONG, HRESULT;
typedef long long LONG_PTR, INT_PTR;
typedef unsigned long long UINT_PTR, ULONG_PTR, SIZE_T;
typedef UINT_PTR WPARAM;
typedef LONG_PTR LPARAM, LRESULT;
typedef void *LPVOID, *PVOID, *HANDLE, *HGDIOBJ;
typedef struct HWND__ *HWND;
typedef struct HINSTANCE__ *HINSTANCE;
typedef HINSTANCE HMODULE;
typedef struct HMENU__ *HMENU;
typedef struct HICON__ *HICON;
typedef HICON HCURSOR;
typedef struct HBRUSH__ *HBRUSH;
typedef struct HFONT__ *HFONT;
typedef struct HDC__ *HDC;
typedef struct HRGN__ *HRGN;
typedef struct HRSRC__ *HRSRC;
typedef struct HMONITOR__ *HMONITOR;
typedef struct DPI_AWARENESS_CONTEXT__ *DPI_AWARENESS_CONTEXT;
typedef LRESULT (*WNDPROC)(HWND, UINT, WPARAM, LPARAM);
typedef void (*TIMERPROC)(HWND, UINT, UINT_PTR, DWORD);
typedef long long (*FARPROC)(void);
typedef BOOL (*WNDENUMPROC)(HWND, LPARAM);
typedef int TASKDIALOG_COMMON_BUTTON_FLAGS, TASKDIALOG_FLAGS;
typedef struct _TASKDIALOG_BUTTON TASKDIALOG_BUTTON;
typedef HRESULT (*PFTASKDIALOGCALLBACK)(HWND, UINT, WPARAM, LPARAM, LONG_PTR);
typedef struct _ITEMIDLIST ITEMIDLIST;
typedef ITEMIDLIST *PIDLIST_ABSOLUTE;
typedef const ITEMIDLIST *PCIDLIST_ABSOLUTE;
typedef int (*BFFCALLBACK)(HWND, UINT, LPARAM, LPARAM);
typedef UINT_PTR (*LPOFNHOOKPROC)(HWND, UINT, WPARAM, LPARAM);

typedef struct tagPOINT { LONG x, y; } POINT;
typedef struct tagRECT { LONG left, top, right, bottom; } RECT;
typedef struct tagSIZE { LONG cx, cy; } SIZE;
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
typedef struct tagMINMAXINFO { POINT ptReserved, ptMaxSize, ptMaxPosition, ptMinTrackSize, ptMaxTrackSize; } MINMAXINFO;
typedef struct tagLOGFONTW {
	LONG lfHeight, lfWidth, lfEscapement, lfOrientation, lfWeight;
	BYTE lfItalic, lfUnderline, lfStrikeOut, lfCharSet, lfOutPrecision, lfClipPrecision, lfQuality, lfPitchAndFamily;
	WCHAR lfFaceName[32];
} LOGFONTW;
typedef struct tagNONCLIENTMETRICSW {
	UINT cbSize;
	int iBorderWidth, iScrollWidth, iScrollHeight, iCaptionWidth, iCaptionHeight;
	LOGFONTW lfCaptionFont;
	int iSmCaptionWidth, iSmCaptionHeight;
	LOGFONTW lfSmCaptionFont;
	int iMenuWidth, iMenuHeight;
	LOGFONTW lfMenuFont, lfStatusFont, lfMessageFont;
	int iPaddedBorderWidth;
} NONCLIENTMETRICSW;
typedef struct tagTEXTMETRICW {
	LONG tmHeight, tmAscent, tmDescent, tmInternalLeading, tmExternalLeading, tmAveCharWidth, tmMaxCharWidth,
		tmWeight, tmOverhang, tmDigitizedAspectX, tmDigitizedAspectY;
	WCHAR tmFirstChar, tmLastChar, tmDefaultChar, tmBreakChar;
	BYTE tmItalic, tmUnderlined, tmStruckOut, tmPitchAndFamily, tmCharSet;
} TEXTMETRICW;
typedef struct tagMONITORINFO { DWORD cbSize; RECT rcMonitor, rcWork; DWORD dwFlags; } MONITORINFO;
typedef struct tagSCROLLINFO { UINT cbSize, fMask; int nMin, nMax; UINT nPage; int nPos, nTrackPos; } SCROLLINFO;
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
typedef struct tagOFNW {
	DWORD lStructSize;
	HWND hwndOwner;
	HINSTANCE hInstance;
	LPCWSTR lpstrFilter;
	LPWSTR lpstrCustomFilter;
	DWORD nMaxCustFilter, nFilterIndex;
	LPWSTR lpstrFile;
	DWORD nMaxFile;
	LPWSTR lpstrFileTitle;
	DWORD nMaxFileTitle;
	LPCWSTR lpstrInitialDir, lpstrTitle;
	DWORD Flags;
	WORD nFileOffset, nFileExtension;
	LPCWSTR lpstrDefExt;
	LPARAM lCustData;
	LPOFNHOOKPROC lpfnHook;
	LPCWSTR lpTemplateName;
	void *pvReserved;
	DWORD dwReserved, FlagsEx;
} OPENFILENAMEW;
#pragma pack(push, 1) /* как в commctrl.h: TASKDIALOGCONFIG упакована */
typedef struct _TASKDIALOGCONFIG {
	UINT cbSize;
	HWND hwndParent;
	HINSTANCE hInstance;
	TASKDIALOG_FLAGS dwFlags;
	TASKDIALOG_COMMON_BUTTON_FLAGS dwCommonButtons;
	PCWSTR pszWindowTitle;
	union { HICON hMainIcon; PCWSTR pszMainIcon; };
	PCWSTR pszMainInstruction, pszContent;
	UINT cButtons;
	const TASKDIALOG_BUTTON *pButtons;
	int nDefaultButton;
	UINT cRadioButtons;
	const TASKDIALOG_BUTTON *pRadioButtons;
	int nDefaultRadioButton;
	PCWSTR pszVerificationText, pszExpandedInformation, pszExpandedControlText, pszCollapsedControlText;
	union { HICON hFooterIcon; PCWSTR pszFooterIcon; };
	PCWSTR pszFooter;
	PFTASKDIALOGCALLBACK pfCallback;
	LONG_PTR lpCallbackData;
	UINT cxWidth;
} TASKDIALOGCONFIG;
#pragma pack(pop)
typedef struct tagDRAWITEMSTRUCT {
	UINT CtlType, CtlID, itemID, itemAction, itemState;
	HWND hwndItem;
	HDC hDC;
	RECT rcItem;
	ULONG_PTR itemData;
} DRAWITEMSTRUCT;
typedef struct tagTOOLINFOW {
	UINT cbSize, uFlags;
	HWND hwnd;
	UINT_PTR uId;
	RECT rect;
	HINSTANCE hinst;
	LPWSTR lpszText;
	LPARAM lParam;
	void *lpReserved;
} TOOLINFOW;
typedef struct _browseinfoW {
	HWND hwndOwner;
	PCIDLIST_ABSOLUTE pidlRoot;
	LPWSTR pszDisplayName;
	LPCWSTR lpszTitle;
	UINT ulFlags;
	BFFCALLBACK lpfn;
	LPARAM lParam;
	int iImage;
} BROWSEINFOW;
typedef struct HBITMAP__ *HBITMAP;
typedef struct HDROP__ *HDROP;
typedef struct tagMENUITEMINFOW {
	UINT cbSize, fMask, fType, fState, wID;
	HMENU hSubMenu;
	HBITMAP hbmpChecked, hbmpUnchecked;
	ULONG_PTR dwItemData;
	LPWSTR dwTypeData;
	UINT cch;
	HBITMAP hbmpItem;
} MENUITEMINFOW;
typedef struct tagCHANGEFILTERSTRUCT CHANGEFILTERSTRUCT;
typedef struct _DROPFILES { DWORD pFiles; POINT pt; BOOL fNC, fWide; } DROPFILES;
/* ListView (List) */
typedef struct tagNMHDR { HWND hwndFrom; UINT_PTR idFrom; UINT code; } NMHDR;
typedef struct tagLVITEMW {
	UINT mask;
	int iItem, iSubItem;
	UINT state, stateMask;
	LPWSTR pszText;
	int cchTextMax, iImage;
	LPARAM lParam;
	int iIndent, iGroupId;
	UINT cColumns;
	UINT *puColumns;
	int *piColFmt;
	int iGroup;
} LVITEMW;
typedef struct tagLVCOLUMNW {
	UINT mask;
	int fmt, cx;
	LPWSTR pszText;
	int cchTextMax, iSubItem, iImage, iOrder, cxMin, cxDefault, cxIdeal;
} LVCOLUMNW;
typedef struct tagLVDISPINFOW { NMHDR hdr; LVITEMW item; } NMLVDISPINFOW;
typedef struct tagTCITEMW {
	UINT mask;
	DWORD dwState, dwStateMask;
	LPWSTR pszText;
	int cchTextMax, iImage;
	LPARAM lParam;
} TCITEMW;
typedef struct tagNMLISTVIEW {
	NMHDR hdr;
	int iItem, iSubItem;
	UINT uNewState, uOldState, uChanged;
	POINT ptAction;
	LPARAM lParam;
} NMLISTVIEW;
typedef HANDLE HGLOBAL;
typedef unsigned int UINT32;
typedef struct _SECURITY_ATTRIBUTES SECURITY_ATTRIBUTES;
typedef struct _OVERLAPPED OVERLAPPED;
typedef struct tagBITMAPINFOHEADER {
	DWORD biSize;
	LONG biWidth, biHeight;
	WORD biPlanes, biBitCount;
	DWORD biCompression, biSizeImage;
	LONG biXPelsPerMeter, biYPelsPerMeter;
	DWORD biClrUsed, biClrImportant;
} BITMAPINFOHEADER;
typedef struct tagRGBQUAD { BYTE rgbBlue, rgbGreen, rgbRed, rgbReserved; } RGBQUAD;
typedef struct tagBITMAPINFO { BITMAPINFOHEADER bmiHeader; RGBQUAD bmiColors[1]; } BITMAPINFO;
typedef struct _BLENDFUNCTION { BYTE BlendOp, BlendFlags, SourceConstantAlpha, AlphaFormat; } BLENDFUNCTION;
/* IStream: драйвер вызывает только Release (третий метод IUnknown). */
typedef struct IStream IStream;
typedef struct IStreamVtbl {
	HRESULT (*QueryInterface)(IStream *, const void *, void **);
	ULONG (*AddRef)(IStream *);
	ULONG (*Release)(IStream *);
} IStreamVtbl;
struct IStream { const IStreamVtbl *lpVtbl; };
/* GDI+, плоский API (gdiplus.h в режиме C): объекты — void, перечисления — int. */
typedef void GpImage, GpBitmap, GpGraphics, GpImageAttributes;
typedef int GpStatus, PixelFormat, InterpolationMode, PixelOffsetMode, WrapMode, GpUnit, RotateFlipType,
            SmoothingMode, GpMatrixOrder;
typedef void GpPen, GpBrush, GpSolidFill;
typedef float REAL;
typedef struct tagPAINTSTRUCT {
	HDC hdc;
	BOOL fErase;
	RECT rcPaint;
	BOOL fRestore, fIncUpdate;
	BYTE rgbReserved[32];
} PAINTSTRUCT;
typedef DWORD ARGB;
typedef void *DrawImageAbort;
typedef struct GdiplusStartupInput {
	UINT32 GdiplusVersion;
	void *DebugEventCallback;
	BOOL SuppressBackgroundThread, SuppressExternalCodecs;
} GdiplusStartupInput;
typedef struct GdiplusStartupOutput GdiplusStartupOutput;
#endif

/* Раскладка структур: в обоих режимах — одни и те же размеры (x86-64). */
_Static_assert(sizeof(MSG) == 48, "MSG");
_Static_assert(sizeof(WNDCLASSEXW) == 80, "WNDCLASSEXW");
_Static_assert(sizeof(MINMAXINFO) == 40, "MINMAXINFO");
_Static_assert(sizeof(LOGFONTW) == 92, "LOGFONTW");
_Static_assert(sizeof(NONCLIENTMETRICSW) == 504, "NONCLIENTMETRICSW");
_Static_assert(sizeof(TEXTMETRICW) == 60, "TEXTMETRICW");
_Static_assert(sizeof(MONITORINFO) == 40, "MONITORINFO");
_Static_assert(sizeof(SCROLLINFO) == 28, "SCROLLINFO");
_Static_assert(sizeof(ACTCTXW) == 56, "ACTCTXW");
_Static_assert(sizeof(OPENFILENAMEW) == 152, "OPENFILENAMEW");
_Static_assert(sizeof(BROWSEINFOW) == 64, "BROWSEINFOW");
_Static_assert(sizeof(TASKDIALOGCONFIG) == 160, "TASKDIALOGCONFIG");
_Static_assert(sizeof(TOOLINFOW) == 72, "TOOLINFOW");
_Static_assert(sizeof(DRAWITEMSTRUCT) == 64, "DRAWITEMSTRUCT");
_Static_assert(sizeof(BITMAPINFO) == 44, "BITMAPINFO");
_Static_assert(sizeof(NMHDR) == 24, "NMHDR");
_Static_assert(sizeof(DROPFILES) == 20, "DROPFILES");
_Static_assert(sizeof(MENUITEMINFOW) == 80, "MENUITEMINFOW");
_Static_assert(sizeof(PAINTSTRUCT) == 72, "PAINTSTRUCT");
_Static_assert(sizeof(LVITEMW) == 88, "LVITEMW");
_Static_assert(sizeof(LVCOLUMNW) == 56, "LVCOLUMNW");
_Static_assert(sizeof(NMLVDISPINFOW) == 112, "NMLVDISPINFOW");
_Static_assert(sizeof(NMLISTVIEW) == 64, "NMLISTVIEW");
_Static_assert(sizeof(TCITEMW) == 40, "TCITEMW");
_Static_assert(sizeof(BLENDFUNCTION) == 4, "BLENDFUNCTION");
_Static_assert(sizeof(GdiplusStartupInput) == 24, "GdiplusStartupInput");

/* ---------- 2. Макросы сверки ---------- */
#ifdef NUI_CHECK
#define NUI_FN(ret, name, params) \
	_Static_assert(__builtin_types_compatible_p(__typeof__(&name), ret (__stdcall *) params), #name);
#define NUI_CONST(name, val) \
	_Static_assert((val) == (name), #name); enum { NUI_##name = (val) };
/* Функции, которые драйвер ищет через GetProcAddress (ADR-16): сверяется тип указателя. */
#define NUI_FNPTR(ret, name, params) \
	typedef ret (__stdcall *nui_##name##_fn) params; \
	_Static_assert(__builtin_types_compatible_p(__typeof__(&name), nui_##name##_fn), #name);
#else
#define NUI_FN(ret, name, params) ret name params;
#define NUI_CONST(name, val) enum { NUI_##name = (val) };
#define NUI_FNPTR(ret, name, params) typedef ret (*nui_##name##_fn) params;
#endif

/* ---------- 3. Константы ---------- */
NUI_CONST(WM_DESTROY, 0x0002)
NUI_CONST(WM_SIZE, 0x0005)
NUI_CONST(WM_CLOSE, 0x0010)
NUI_CONST(WM_GETMINMAXINFO, 0x0024)
NUI_CONST(WM_SETFONT, 0x0030)
NUI_CONST(WM_GETFONT, 0x0031)
NUI_CONST(WM_ERASEBKGND, 0x0014)
NUI_CONST(WM_DRAWITEM, 0x002B)
NUI_CONST(WM_CTLCOLORBTN, 0x0135)
NUI_CONST(WM_SETICON, 0x0080)
NUI_CONST(WM_KEYDOWN, 0x0100)
NUI_CONST(WM_SYSKEYDOWN, 0x0104)
NUI_CONST(WM_CHAR, 0x0102)
NUI_CONST(VK_SHIFT, 0x10)
NUI_CONST(VK_CONTROL, 0x11)
NUI_CONST(VK_MENU, 0x12)
NUI_CONST(WM_KEYFIRST, 0x0100)
NUI_CONST(WM_KEYLAST, 0x0109)
NUI_CONST(WM_COMMAND, 0x0111)
NUI_CONST(WM_TIMER, 0x0113)
NUI_CONST(WM_HSCROLL, 0x0114)
NUI_CONST(WM_VSCROLL, 0x0115)
NUI_CONST(WM_CTLCOLOREDIT, 0x0133)
NUI_CONST(WM_CTLCOLORSTATIC, 0x0138)
NUI_CONST(WM_MOUSEWHEEL, 0x020A)
NUI_CONST(WM_DPICHANGED, 0x02E0)
NUI_CONST(WM_APP, 0x8000)
NUI_CONST(DM_GETDEFID, 0x0400)
NUI_CONST(DC_HASDEFID, 0x534B)
NUI_CONST(VK_RETURN, 0x0D)

NUI_CONST(WS_POPUP, (DWORD)0x80000000)
NUI_CONST(WS_CHILD, 0x40000000)
NUI_CONST(WS_VISIBLE, 0x10000000)
NUI_CONST(WS_CLIPSIBLINGS, 0x04000000)
NUI_CONST(WS_CLIPCHILDREN, 0x02000000)
NUI_CONST(WS_OVERLAPPEDWINDOW, 0x00CF0000)
NUI_CONST(WS_THICKFRAME, 0x00040000)
NUI_CONST(WS_MAXIMIZEBOX, 0x00010000)
NUI_CONST(WS_VSCROLL, 0x00200000)
NUI_CONST(WS_HSCROLL, 0x00100000)
NUI_CONST(WS_TABSTOP, 0x00010000)
NUI_CONST(WS_GROUP, 0x00020000)
NUI_CONST(WS_EX_CONTROLPARENT, 0x00010000)
NUI_CONST(WS_EX_CLIENTEDGE, 0x00000200)
NUI_CONST(GWL_STYLE, -16)
NUI_CONST(GWL_EXSTYLE, -20)
NUI_CONST(GWLP_USERDATA, -21)
NUI_CONST(GWLP_ID, -12)
NUI_CONST(GA_ROOT, 2)
NUI_CONST(SW_HIDE, 0)
NUI_CONST(SW_SHOWNORMAL, 1)
NUI_CONST(SW_SHOW, 5)
NUI_CONST(SW_SHOWNA, 8)
NUI_CONST(SW_MAXIMIZE, 3)
NUI_CONST(SW_SHOWMAXIMIZED, 3)
NUI_CONST(SW_RESTORE, 9)
NUI_CONST(SWP_NOSIZE, 0x0001)
NUI_CONST(SWP_NOMOVE, 0x0002)
NUI_CONST(SWP_NOZORDER, 0x0004)
NUI_CONST(SWP_NOACTIVATE, 0x0010)
NUI_CONST(SWP_FRAMECHANGED, 0x0020)
NUI_CONST(SWP_NOOWNERZORDER, 0x0200)
NUI_CONST(CW_USEDEFAULT, (int)0x80000000)
NUI_CONST(MONITOR_DEFAULTTONEAREST, 2)
NUI_CONST(SPI_GETNONCLIENTMETRICS, 0x0029)
NUI_CONST(COLOR_WINDOW, 5)
NUI_CONST(COLOR_WINDOWTEXT, 8)
NUI_CONST(COLOR_BTNFACE, 15)
NUI_CONST(COLOR_GRAYTEXT, 17)
NUI_CONST(COLOR_BTNTEXT, 18)
NUI_CONST(RDW_INVALIDATE, 0x0001)
NUI_CONST(RDW_ERASE, 0x0004)
NUI_CONST(RDW_ALLCHILDREN, 0x0080)
NUI_CONST(ODS_SELECTED, 0x0001)
NUI_CONST(ODS_DISABLED, 0x0004)
NUI_CONST(ODS_FOCUS, 0x0010)
NUI_CONST(ODS_NOFOCUSRECT, 0x0200)
NUI_CONST(DT_CENTER, 0x00000001)
NUI_CONST(DT_VCENTER, 0x00000004)
NUI_CONST(SM_CXVSCROLL, 2)
NUI_CONST(SM_CYHSCROLL, 3)
NUI_CONST(SM_CXEDGE, 45)
NUI_CONST(SM_CYEDGE, 46)
NUI_CONST(SM_CXMENUCHECK, 71)
NUI_CONST(SM_CYMENUCHECK, 72)
NUI_CONST(FW_NORMAL, 400)
NUI_CONST(FW_BOLD, 700)
NUI_CONST(TRANSPARENT, 1)
NUI_CONST(DT_SINGLELINE, 0x00000020)
NUI_CONST(DT_WORDBREAK, 0x00000010)
NUI_CONST(DT_CALCRECT, 0x00000400)
NUI_CONST(DT_NOPREFIX, 0x00000800)
NUI_CONST(DT_EXPANDTABS, 0x00000040)
NUI_CONST(DT_EDITCONTROL, 0x00002000)
NUI_CONST(GWLP_WNDPROC, -4)
NUI_CONST(WM_PAINT, 0x000F)
NUI_CONST(WM_NCDESTROY, 0x0082)
NUI_CONST(WM_SETTEXT, 0x000C)
NUI_CONST(WM_CUT, 0x0300)
NUI_CONST(WM_PASTE, 0x0302)
NUI_CONST(WM_CLEAR, 0x0303)
NUI_CONST(WM_UNDO, 0x0304)
NUI_CONST(EM_UNDO, 0x00C7)
NUI_CONST(EM_GETRECT, 0x00B2)
NUI_CONST(ICC_STANDARD_CLASSES, 0x00004000)
NUI_CONST(ICC_BAR_CLASSES, 0x00000004)
NUI_CONST(ICC_LISTVIEW_CLASSES, 0x00000001)
NUI_CONST(WM_NOTIFY, 0x004E)
NUI_CONST(WM_DROPFILES, 0x0233)
NUI_CONST(MF_STRING, 0x0000)
NUI_CONST(MF_GRAYED, 0x0001)
NUI_CONST(MF_CHECKED, 0x0008)
NUI_CONST(MF_POPUP, 0x0010)
NUI_CONST(MF_BYPOSITION, 0x0400)
NUI_CONST(MF_SEPARATOR, 0x0800)
NUI_CONST(MIIM_STRING, 0x0040)
NUI_CONST(WM_COPYDATA, 0x004A)
NUI_CONST(MSGFLT_ALLOW, 1)
NUI_CONST(LVS_REPORT, 0x0001)
NUI_CONST(LVS_SINGLESEL, 0x0004)
NUI_CONST(LVS_SHOWSELALWAYS, 0x0008)
NUI_CONST(LVS_OWNERDATA, 0x1000)
NUI_CONST(LVS_NOCOLUMNHEADER, 0x4000)
NUI_CONST(LVS_EX_FULLROWSELECT, 0x00000020)
NUI_CONST(LVS_EX_LABELTIP, 0x00004000)
NUI_CONST(LVS_EX_DOUBLEBUFFER, 0x00010000)
NUI_CONST(LVM_SETEXTENDEDLISTVIEWSTYLE, 0x1036)
NUI_CONST(ICC_TAB_CLASSES, 0x00000008)
NUI_CONST(TCIF_TEXT, 0x0001)
NUI_CONST(TCM_DELETEALLITEMS, 0x1309)
NUI_CONST(TCM_GETITEMRECT, 0x130A)
NUI_CONST(TCM_GETCURSEL, 0x130B)
NUI_CONST(TCM_SETCURSEL, 0x130C)
NUI_CONST(TCM_ADJUSTRECT, 0x1328)
NUI_CONST(TCM_INSERTITEMW, 0x133E)
/* TCN_SELCHANGE = TCN_FIRST - 1 = 0U - 551U: в enum не помещается (TinyGo сделал бы его
   отрицательным) — в Go это константа tcnSelChange, здесь только сверка. */
#ifdef NUI_CHECK
_Static_assert(TCN_SELCHANGE == 0xFFFFFDD9u, "TCN_SELCHANGE");
#endif
NUI_CONST(WM_PRINTCLIENT, 0x0318)
NUI_CONST(PRF_CLIENT, 0x00000004)
NUI_CONST(PRF_ERASEBKGND, 0x00000008)
NUI_CONST(RGN_DIFF, 4)
NUI_CONST(SM_CXHSCROLL, 21)
NUI_CONST(VK_TAB, 0x09)
NUI_CONST(VK_PRIOR, 0x21)
NUI_CONST(VK_NEXT, 0x22)
NUI_CONST(VK_LEFT, 0x25)
NUI_CONST(VK_UP, 0x26)
NUI_CONST(VK_DOWN, 0x28)
NUI_CONST(BS_RADIOBUTTON, 0x0004)
NUI_CONST(WM_GETDLGCODE, 0x0087)
NUI_CONST(DLGC_RADIOBUTTON, 0x0040)
NUI_CONST(LVM_INSERTCOLUMNW, 0x1061)
NUI_CONST(LVM_SETCOLUMNWIDTH, 0x101E)
NUI_CONST(LVM_GETCOLUMNWIDTH, 0x101D)
NUI_CONST(LVM_SETITEMCOUNT, 0x102F)
NUI_CONST(LVM_GETNEXTITEM, 0x100C)
NUI_CONST(LVM_SETITEMSTATE, 0x102B)
NUI_CONST(LVM_ENSUREVISIBLE, 0x1013)
NUI_CONST(LVM_GETSTRINGWIDTHW, 0x1057)
NUI_CONST(LVM_GETITEMTEXTW, 0x1073)
NUI_CONST(LVCF_WIDTH, 0x0002)
NUI_CONST(LVCF_TEXT, 0x0004)
NUI_CONST(LVCF_SUBITEM, 0x0008)
NUI_CONST(LVNI_SELECTED, 0x0002)
NUI_CONST(LVIS_FOCUSED, 0x0001)
NUI_CONST(LVIS_SELECTED, 0x0002)
NUI_CONST(LVIF_TEXT, 0x0001)
NUI_CONST(LVIF_STATE, 0x0008)
NUI_CONST(LVN_ITEMCHANGED, -101)
NUI_CONST(LVN_ITEMACTIVATE, -114)
NUI_CONST(LVN_GETDISPINFOW, -177)
NUI_CONST(ICC_PROGRESS_CLASS, 0x00000020)
NUI_CONST(ACTCTX_FLAG_ASSEMBLY_DIRECTORY_VALID, 0x004)
NUI_CONST(ACTCTX_FLAG_RESOURCE_NAME_VALID, 0x008)
NUI_CONST(ACTCTX_FLAG_SET_PROCESS_DEFAULT, 0x010)

/* Static, Button, Edit, ComboBox, Trackbar, Progress */
NUI_CONST(SS_LEFT, 0x0000)
NUI_CONST(SS_CENTER, 0x0001)
NUI_CONST(SS_RIGHT, 0x0002)
NUI_CONST(SS_LEFTNOWORDWRAP, 0x000C)
NUI_CONST(SS_ETCHEDHORZ, 0x0010)
NUI_CONST(SS_ETCHEDVERT, 0x0011)
NUI_CONST(SS_NOPREFIX, 0x0080)
NUI_CONST(SS_CENTERIMAGE, 0x0200)
NUI_CONST(SS_OWNERDRAW, 0x000D)
NUI_CONST(SS_ENDELLIPSIS, 0x4000)
NUI_CONST(SS_PATHELLIPSIS, 0x8000)
NUI_CONST(BS_PUSHBUTTON, 0x0000)
NUI_CONST(BS_DEFPUSHBUTTON, 0x0001)
NUI_CONST(BS_AUTOCHECKBOX, 0x0003)
NUI_CONST(BS_GROUPBOX, 0x0007)
NUI_CONST(BS_OWNERDRAW, 0x000B)
NUI_CONST(BM_GETCHECK, 0x00F0)
NUI_CONST(BM_SETCHECK, 0x00F1)
NUI_CONST(BM_SETSTYLE, 0x00F4)
NUI_CONST(BM_CLICK, 0x00F5)
NUI_CONST(BCM_GETIDEALSIZE, 0x1601)
NUI_CONST(BN_CLICKED, 0)
NUI_CONST(ES_MULTILINE, 0x0004)
NUI_CONST(ES_AUTOVSCROLL, 0x0040)
NUI_CONST(ES_AUTOHSCROLL, 0x0080)
NUI_CONST(ES_NOHIDESEL, 0x0100)
NUI_CONST(ES_READONLY, 0x0800)
NUI_CONST(ES_WANTRETURN, 0x1000)
NUI_CONST(EN_CHANGE, 0x0300)
NUI_CONST(EN_KILLFOCUS, 0x0200)
NUI_CONST(ES_NUMBER, 0x2000)
NUI_CONST(ICC_UPDOWN_CLASS, 0x00000010)
NUI_CONST(UDS_SETBUDDYINT, 0x0002)
NUI_CONST(UDS_ARROWKEYS, 0x0020)
NUI_CONST(UDS_NOTHOUSANDS, 0x0080)
NUI_CONST(UDM_SETBUDDY, 0x0469)
NUI_CONST(UDM_SETRANGE32, 0x046F)
NUI_CONST(UDM_SETPOS32, 0x0471)
NUI_CONST(WM_KEYUP, 0x0101)
NUI_CONST(EM_GETSEL, 0x00B0)
NUI_CONST(EM_SETSEL, 0x00B1)
NUI_CONST(EM_LINESCROLL, 0x00B6)
NUI_CONST(EM_GETLINECOUNT, 0x00BA)
NUI_CONST(EM_LINEINDEX, 0x00BB)
NUI_CONST(EM_REPLACESEL, 0x00C2)
NUI_CONST(EM_SETLIMITTEXT, 0x00C5)
NUI_CONST(EM_SETPASSWORDCHAR, 0x00CC)
NUI_CONST(EM_GETFIRSTVISIBLELINE, 0x00CE)
NUI_CONST(EM_SETREADONLY, 0x00CF)
NUI_CONST(EM_SETCUEBANNER, 0x1501)
NUI_CONST(EM_SETMARGINS, 0x00D3)
NUI_CONST(EC_LEFTMARGIN, 0x0001)
NUI_CONST(EC_RIGHTMARGIN, 0x0002)
NUI_CONST(ES_CENTER, 0x0001)
NUI_CONST(ES_RIGHT, 0x0002)
NUI_CONST(TTM_SETMAXTIPWIDTH, 0x0418)
NUI_CONST(TTM_ADDTOOLW, 0x0432)
NUI_CONST(TTM_DELTOOLW, 0x0433)
NUI_CONST(TTM_UPDATETIPTEXTW, 0x0439)
NUI_CONST(TTF_IDISHWND, 0x0001)
NUI_CONST(TTF_SUBCLASS, 0x0010)
NUI_CONST(TTS_ALWAYSTIP, 0x01)
NUI_CONST(TTS_NOPREFIX, 0x02)
NUI_CONST(WS_EX_TOPMOST, 0x00000008)
NUI_CONST(CBS_DROPDOWN, 0x0002)
NUI_CONST(CBS_DROPDOWNLIST, 0x0003)
NUI_CONST(CBS_AUTOHSCROLL, 0x0040)
NUI_CONST(CB_ADDSTRING, 0x0143)
NUI_CONST(CB_GETCOUNT, 0x0146)
NUI_CONST(CB_GETCURSEL, 0x0147)
NUI_CONST(CB_GETLBTEXT, 0x0148)
NUI_CONST(CB_GETLBTEXTLEN, 0x0149)
NUI_CONST(CB_RESETCONTENT, 0x014B)
NUI_CONST(CB_SETCURSEL, 0x014E)
NUI_CONST(CB_GETITEMHEIGHT, 0x0154)
NUI_CONST(CB_SETDROPPEDWIDTH, 0x0160)
NUI_CONST(CBN_SELCHANGE, 1)
NUI_CONST(CBN_EDITCHANGE, 5)
NUI_CONST(TBS_NOTICKS, 0x0010)
NUI_CONST(TBM_GETPOS, 0x0400)
NUI_CONST(TBM_SETPOS, 0x0405)
NUI_CONST(TBM_SETRANGEMIN, 0x0407)
NUI_CONST(TBM_SETRANGEMAX, 0x0408)
NUI_CONST(PBS_MARQUEE, 0x08)
NUI_CONST(PBM_SETPOS, 0x0402)
NUI_CONST(PBM_SETRANGE32, 0x0406)
NUI_CONST(PBM_SETMARQUEE, 0x040A)

/* Прокрутка */
NUI_CONST(SB_HORZ, 0)
NUI_CONST(SB_VERT, 1)
NUI_CONST(SIF_RANGE, 0x0001)
NUI_CONST(SIF_PAGE, 0x0002)
NUI_CONST(SIF_POS, 0x0004)
NUI_CONST(SIF_TRACKPOS, 0x0010)
NUI_CONST(SB_LINEUP, 0)
NUI_CONST(SB_LINEDOWN, 1)
NUI_CONST(SB_PAGEUP, 2)
NUI_CONST(SB_PAGEDOWN, 3)
NUI_CONST(SB_THUMBTRACK, 5)
NUI_CONST(SB_TOP, 6)
NUI_CONST(SB_BOTTOM, 7)

/* Диалоги и оболочка */
NUI_CONST(MB_OK, 0x0)
NUI_CONST(MB_YESNO, 0x4)
NUI_CONST(MB_ICONERROR, 0x10)
NUI_CONST(MB_ICONQUESTION, 0x20)
NUI_CONST(MB_ICONINFORMATION, 0x40)
NUI_CONST(IDOK, 1)
NUI_CONST(IDYES, 6)
NUI_CONST(TDCBF_OK_BUTTON, 0x1)
NUI_CONST(TDCBF_YES_BUTTON, 0x2)
NUI_CONST(TDCBF_NO_BUTTON, 0x4)
NUI_CONST(TDF_ALLOW_DIALOG_CANCELLATION, 0x0008)
NUI_CONST(TDF_POSITION_RELATIVE_TO_WINDOW, 0x1000)
NUI_CONST(TDM_CLICK_BUTTON, 0x0466)
NUI_CONST(IDCANCEL, 2)
NUI_CONST(IDNO, 7)
NUI_CONST(GW_CHILD, 5)
NUI_CONST(OFN_OVERWRITEPROMPT, 0x00000002)
NUI_CONST(OFN_HIDEREADONLY, 0x00000004)
NUI_CONST(OFN_ALLOWMULTISELECT, 0x00000200)
NUI_CONST(OFN_NOCHANGEDIR, 0x00000008)
NUI_CONST(OFN_PATHMUSTEXIST, 0x00000800)
NUI_CONST(OFN_FILEMUSTEXIST, 0x00001000)
NUI_CONST(OFN_EXPLORER, 0x00080000)
NUI_CONST(BIF_RETURNONLYFSDIRS, 0x00000001)
NUI_CONST(BIF_NEWDIALOGSTYLE, 0x00000040)
NUI_CONST(COINIT_APARTMENTTHREADED, 0x2)
NUI_CONST(CF_UNICODETEXT, 13)
NUI_CONST(GMEM_MOVEABLE, 0x0002)
NUI_CONST(GENERIC_READ, 0x80000000)
NUI_CONST(FILE_SHARE_READ, 0x00000001)
NUI_CONST(OPEN_EXISTING, 3)
NUI_CONST(FILE_ATTRIBUTE_NORMAL, 0x80)
NUI_CONST(BI_RGB, 0)
NUI_CONST(DIB_RGB_COLORS, 0)
NUI_CONST(AC_SRC_OVER, 0x00)
NUI_CONST(AC_SRC_ALPHA, 0x01)
/* GDI+ */
NUI_CONST(PixelFormat32bppPARGB, 0xE200B)
NUI_CONST(InterpolationModeHighQualityBicubic, 7)
NUI_CONST(PixelOffsetModeHighQuality, 2)
NUI_CONST(PixelOffsetModeHalf, 4)
NUI_CONST(SmoothingModeAntiAlias, 4)
NUI_CONST(MatrixOrderPrepend, 0)
NUI_CONST(UnitWorld, 0)
NUI_CONST(SRCCOPY, 0x00CC0020)
NUI_CONST(WM_MOUSEMOVE, 0x0200)
NUI_CONST(WM_SETCURSOR, 0x0020)
NUI_CONST(WM_CONTEXTMENU, 0x007B)
NUI_CONST(TPM_RIGHTBUTTON, 0x0002)
NUI_CONST(TPM_NONOTIFY, 0x0080)
NUI_CONST(TPM_RETURNCMD, 0x0100)
NUI_CONST(WM_LBUTTONDOWN, 0x0201)
NUI_CONST(WM_LBUTTONUP, 0x0202)
NUI_CONST(WM_RBUTTONDOWN, 0x0204)
NUI_CONST(WM_RBUTTONUP, 0x0205)
NUI_CONST(WM_MBUTTONDOWN, 0x0207)
NUI_CONST(WM_MBUTTONUP, 0x0208)
NUI_CONST(WM_MOUSEHWHEEL, 0x020E)
NUI_CONST(MK_SHIFT, 0x0004)
NUI_CONST(MK_CONTROL, 0x0008)
NUI_CONST(WrapModeTileFlipXY, 3)
NUI_CONST(UnitPixel, 2)

/* ---------- 4. Функции ---------- */
/* kernel32 */
NUI_FN(HMODULE, GetModuleHandleW, (LPCWSTR lpModuleName))
NUI_FN(HMODULE, LoadLibraryW, (LPCWSTR lpLibFileName))
NUI_FN(FARPROC, GetProcAddress, (HMODULE hModule, const char *lpProcName))
NUI_FN(DWORD, GetCurrentThreadId, (void))
NUI_FN(DWORD, GetModuleFileNameW, (HMODULE hModule, LPWSTR lpFilename, DWORD nSize))
NUI_FN(DWORD, GetLastError, (void))
NUI_FN(UINT, GetSystemDirectoryW, (LPWSTR lpBuffer, UINT uSize))
NUI_FN(HANDLE, CreateActCtxW, (const ACTCTXW *pActCtx))
NUI_FN(BOOL, ActivateActCtx, (HANDLE hActCtx, ULONG_PTR *lpCookie))
NUI_FN(HRSRC, FindResourceW, (HMODULE hModule, LPCWSTR lpName, LPCWSTR lpType))
NUI_FN(HANDLE, GlobalAlloc, (UINT uFlags, SIZE_T dwBytes))
NUI_FN(LPVOID, GlobalLock, (HANDLE hMem))
NUI_FN(BOOL, GlobalUnlock, (HANDLE hMem))
NUI_FN(HANDLE, GlobalFree, (HANDLE hMem))
NUI_FN(void, Sleep, (DWORD dwMilliseconds))
NUI_FN(HANDLE, CreateFileW, (LPCWSTR lpFileName, DWORD dwDesiredAccess, DWORD dwShareMode,
       SECURITY_ATTRIBUTES *lpSecurityAttributes, DWORD dwCreationDisposition, DWORD dwFlagsAndAttributes,
       HANDLE hTemplateFile))
NUI_FN(DWORD, GetFileSize, (HANDLE hFile, DWORD *lpFileSizeHigh))
NUI_FN(BOOL, ReadFile, (HANDLE hFile, LPVOID lpBuffer, DWORD nNumberOfBytesToRead, DWORD *lpNumberOfBytesRead,
       OVERLAPPED *lpOverlapped))
NUI_FN(BOOL, CloseHandle, (HANDLE hObject))

/* user32: окна и сообщения */
NUI_FN(ATOM, RegisterClassExW, (const WNDCLASSEXW *wc))
NUI_FN(HWND, CreateWindowExW, (DWORD dwExStyle, LPCWSTR lpClassName, LPCWSTR lpWindowName, DWORD dwStyle,
       int X, int Y, int nWidth, int nHeight, HWND hWndParent, HMENU hMenu, HINSTANCE hInstance, LPVOID lpParam))
NUI_FN(BOOL, DestroyWindow, (HWND hWnd))
NUI_FN(LRESULT, DefWindowProcW, (HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam))
NUI_FN(BOOL, GetMessageW, (MSG *lpMsg, HWND hWnd, UINT wMsgFilterMin, UINT wMsgFilterMax))
NUI_FN(BOOL, TranslateMessage, (const MSG *lpMsg))
NUI_FN(LRESULT, DispatchMessageW, (const MSG *lpMsg))
NUI_FN(BOOL, IsDialogMessageW, (HWND hDlg, MSG *lpMsg))
NUI_FN(void, PostQuitMessage, (int nExitCode))
NUI_FN(BOOL, PostMessageW, (HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam))
NUI_FN(LRESULT, SendMessageW, (HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam))
NUI_FN(BOOL, ShowWindow, (HWND hWnd, int nCmdShow))
NUI_FN(BOOL, IsWindowVisible, (HWND hWnd))
NUI_FN(BOOL, EnableWindow, (HWND hWnd, BOOL bEnable))
NUI_FN(BOOL, SetWindowTextW, (HWND hWnd, LPCWSTR lpString))
NUI_FN(int, GetWindowTextW, (HWND hWnd, LPWSTR lpString, int nMaxCount))
NUI_FN(int, GetWindowTextLengthW, (HWND hWnd))
NUI_FN(LONG_PTR, SetWindowLongPtrW, (HWND hWnd, int nIndex, LONG_PTR dwNewLong))
NUI_FN(LONG_PTR, GetWindowLongPtrW, (HWND hWnd, int nIndex))
NUI_FN(BOOL, SetPropW, (HWND hWnd, LPCWSTR lpString, HANDLE hData))
NUI_FN(HANDLE, GetPropW, (HWND hWnd, LPCWSTR lpString))
NUI_FN(LRESULT, CallWindowProcW, (WNDPROC lpPrevWndFunc, HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam))
NUI_FN(SHORT, GetKeyState, (int nVirtKey))
NUI_FN(BOOL, ChangeWindowMessageFilterEx, (HWND hwnd, UINT message, DWORD action, CHANGEFILTERSTRUCT *pChangeFilterStruct))
NUI_FN(HWND, GetFocus, (void))
NUI_FN(HANDLE, RemovePropW, (HWND hWnd, LPCWSTR lpString))
NUI_FN(HWND, GetAncestor, (HWND hwnd, UINT gaFlags))
NUI_FN(HWND, SetParent, (HWND hWndChild, HWND hWndNewParent))
NUI_FN(BOOL, SetWindowPos, (HWND hWnd, HWND hWndInsertAfter, int X, int Y, int cx, int cy, UINT uFlags))
NUI_FN(HMENU, CreateMenu, (void))
NUI_FN(HMENU, CreatePopupMenu, (void))
NUI_FN(BOOL, TrackPopupMenu, (HMENU hMenu, UINT uFlags, int x, int y, int nReserved, HWND hWnd, const RECT *prcRect))
NUI_FN(BOOL, AppendMenuW, (HMENU hMenu, UINT uFlags, UINT_PTR uIDNewItem, LPCWSTR lpNewItem))
NUI_FN(BOOL, SetMenu, (HWND hWnd, HMENU hMenu))
NUI_FN(HMENU, GetMenu, (HWND hWnd))
NUI_FN(BOOL, DrawMenuBar, (HWND hWnd))
NUI_FN(BOOL, DestroyMenu, (HMENU hMenu))
NUI_FN(BOOL, EnableMenuItem, (HMENU hMenu, UINT uIDEnableItem, UINT uEnable))
NUI_FN(DWORD, CheckMenuItem, (HMENU hmenu, UINT uIDCheckItem, UINT uCheck))
NUI_FN(UINT, GetMenuState, (HMENU hMenu, UINT uId, UINT uFlags))
NUI_FN(int, GetMenuItemCount, (HMENU hMenu))
NUI_FN(BOOL, SetMenuItemInfoW, (HMENU hmenu, UINT item, BOOL fByPositon, const MENUITEMINFOW *lpmii))
NUI_FN(int, GetMenuStringW, (HMENU hMenu, UINT uIDItem, LPWSTR lpString, int cchMax, UINT flags))
NUI_FN(BOOL, GetClientRect, (HWND hWnd, RECT *lpRect))
NUI_FN(HDC, BeginPaint, (HWND hWnd, PAINTSTRUCT *lpPaint))
NUI_FN(BOOL, EndPaint, (HWND hWnd, const PAINTSTRUCT *lpPaint))
NUI_FN(HBITMAP, CreateCompatibleBitmap, (HDC hdc, int cx, int cy))
NUI_FN(BOOL, BitBlt, (HDC hdc, int x, int y, int cx, int cy, HDC hdcSrc, int x1, int y1, DWORD rop))
NUI_FN(HWND, SetCapture, (HWND hWnd))
NUI_FN(BOOL, ReleaseCapture, (void))
NUI_FN(BOOL, ScreenToClient, (HWND hWnd, POINT *lpPoint))
NUI_FN(BOOL, ClientToScreen, (HWND hWnd, POINT *lpPoint))                                   /* тесты */
NUI_FN(COLORREF, GetPixel, (HDC hdc, int x, int y))                                         /* тесты */
NUI_FN(BOOL, GetWindowRect, (HWND hWnd, RECT *lpRect))
NUI_FN(HWND, SetFocus, (HWND hWnd))
NUI_FN(BOOL, SetForegroundWindow, (HWND hWnd))
NUI_FN(BOOL, InvalidateRect, (HWND hWnd, const RECT *lpRect, BOOL bErase))
NUI_FN(UINT_PTR, SetTimer, (HWND hWnd, UINT_PTR nIDEvent, UINT uElapse, TIMERPROC lpTimerFunc))
NUI_FN(BOOL, KillTimer, (HWND hWnd, UINT_PTR uIDEvent))
NUI_FN(HCURSOR, LoadCursorW, (HINSTANCE hInstance, LPCWSTR lpCursorName))
NUI_FN(HCURSOR, SetCursor, (HCURSOR hCursor))
NUI_FN(HWND, GetCapture, (void))
NUI_FN(BOOL, GetCursorPos, (POINT *lpPoint))
NUI_FN(BOOL, SetCursorPos, (int X, int Y))                                                  /* тесты */
NUI_FN(HICON, LoadIconW, (HINSTANCE hInstance, LPCWSTR lpIconName))
NUI_FN(DWORD, GetSysColor, (int nIndex))
NUI_FN(HBRUSH, GetSysColorBrush, (int nIndex))
NUI_FN(HMONITOR, MonitorFromWindow, (HWND hwnd, DWORD dwFlags))
NUI_FN(BOOL, GetMonitorInfoW, (HMONITOR hMonitor, MONITORINFO *lpmi))
NUI_FN(int, MessageBoxW, (HWND hWnd, LPCWSTR lpText, LPCWSTR lpCaption, UINT uType))
NUI_FN(int, SetScrollInfo, (HWND hwnd, int nBar, const SCROLLINFO *lpsi, BOOL redraw))
NUI_FN(BOOL, GetScrollInfo, (HWND hwnd, int nBar, SCROLLINFO *lpsi))
NUI_FN(BOOL, OpenClipboard, (HWND hWndNewOwner))
NUI_FN(BOOL, CloseClipboard, (void))
NUI_FN(BOOL, EmptyClipboard, (void))
NUI_FN(HANDLE, SetClipboardData, (UINT uFormat, HANDLE hMem))
NUI_FN(HANDLE, GetClipboardData, (UINT uFormat))
NUI_FN(HWND, GetWindow, (HWND hWnd, UINT uCmd))
NUI_FN(HWND, GetParent, (HWND hWnd))
NUI_FN(HWND, GetDlgItem, (HWND hDlg, int nIDDlgItem))
NUI_FN(int, GetClassNameW, (HWND hWnd, LPWSTR lpClassName, int nMaxCount))
NUI_FN(BOOL, EnumThreadWindows, (DWORD dwThreadId, WNDENUMPROC lpfn, LPARAM lParam))

/* user32: DPI (Windows 10 1607+, есть и в Wine 8) */
NUI_FN(BOOL, SetProcessDpiAwarenessContext, (DPI_AWARENESS_CONTEXT value))
NUI_FN(UINT, GetDpiForWindow, (HWND hwnd))
NUI_FN(UINT, GetDpiForSystem, (void))
NUI_FN(int, GetSystemMetricsForDpi, (int nIndex, UINT dpi))
NUI_FN(BOOL, IsZoomed, (HWND hWnd))
NUI_FN(HRGN, CreateRectRgn, (int x1, int y1, int x2, int y2))
NUI_FN(int, CombineRgn, (HRGN hrgnDst, HRGN hrgnSrc1, HRGN hrgnSrc2, int iMode))
NUI_FN(int, SetWindowRgn, (HWND hWnd, HRGN hRgn, BOOL bRedraw))
NUI_FN(HWND, WindowFromPoint, (POINT Point))                                               /* тесты */
NUI_FN(BOOL, SystemParametersInfoForDpi, (UINT uiAction, UINT uiParam, PVOID pvParam, UINT fWinIni, UINT dpi))
NUI_FN(BOOL, AdjustWindowRectExForDpi, (RECT *lpRect, DWORD dwStyle, BOOL bMenu, DWORD dwExStyle, UINT dpi))

/* gdi32 */
NUI_FN(HDC, GetDC, (HWND hWnd))
NUI_FN(int, ReleaseDC, (HWND hWnd, HDC hDC))
NUI_FN(HFONT, CreateFontIndirectW, (const LOGFONTW *lplf))
NUI_FN(HGDIOBJ, SelectObject, (HDC hdc, HGDIOBJ h))
NUI_FN(BOOL, DeleteObject, (HGDIOBJ ho))
NUI_FN(BOOL, GetTextMetricsW, (HDC hdc, TEXTMETRICW *lptm))
NUI_FN(BOOL, GetTextExtentPoint32W, (HDC hdc, LPCWSTR lpString, int c, SIZE *psizl))
NUI_FN(int, DrawTextW, (HDC hdc, LPCWSTR lpchText, int cchText, RECT *lprc, UINT format))
NUI_FN(COLORREF, SetTextColor, (HDC hdc, COLORREF color))
NUI_FN(COLORREF, SetBkColor, (HDC hdc, COLORREF color))
NUI_FN(int, SetBkMode, (HDC hdc, int mode))
NUI_FN(HBRUSH, CreateSolidBrush, (COLORREF color))
NUI_FN(int, GetObjectW, (HANDLE h, int c, LPVOID pv))
NUI_FN(int, FillRect, (HDC hDC, const RECT *lprc, HBRUSH hbr))
NUI_FN(HBITMAP, CreateDIBSection, (HDC hdc, const BITMAPINFO *pbmi, UINT usage, void **ppvBits, HANDLE hSection,
       DWORD offset))
NUI_FN(HDC, CreateCompatibleDC, (HDC hdc))
NUI_FN(BOOL, DeleteDC, (HDC hdc))
NUI_FN(BOOL, GdiAlphaBlend, (HDC hdcDest, int xoriginDest, int yoriginDest, int wDest, int hDest, HDC hdcSrc,
       int xoriginSrc, int yoriginSrc, int wSrc, int hSrc, BLENDFUNCTION ftn))
NUI_FN(BOOL, GdiFlush, (void))
NUI_FN(int, FrameRect, (HDC hDC, const RECT *lprc, HBRUSH hbr))
NUI_FN(BOOL, DrawFocusRect, (HDC hDC, const RECT *lprc))
NUI_FN(BOOL, RedrawWindow, (HWND hWnd, const RECT *lprcUpdate, HRGN hrgnUpdate, UINT flags))

/* shell32, ole32, comdlg32 */
NUI_FN(HINSTANCE, ShellExecuteW, (HWND hwnd, LPCWSTR lpOperation, LPCWSTR lpFile, LPCWSTR lpParameters,
       LPCWSTR lpDirectory, INT nShowCmd))
NUI_FN(PIDLIST_ABSOLUTE, SHBrowseForFolderW, (BROWSEINFOW *lpbi))
NUI_FN(BOOL, SHGetPathFromIDListW, (PCIDLIST_ABSOLUTE pidl, LPWSTR pszPath))
NUI_FN(HRESULT, CoInitializeEx, (LPVOID pvReserved, DWORD dwCoInit))
NUI_FN(void, CoTaskMemFree, (LPVOID pv))
NUI_FN(LPVOID, CoTaskMemAlloc, (SIZE_T cb))
NUI_FN(HRESULT, CreateStreamOnHGlobal, (HGLOBAL hGlobal, BOOL fDeleteOnRelease, IStream **ppstm))
NUI_FN(void, DragAcceptFiles, (HWND hWnd, BOOL fAccept))
NUI_FN(UINT, DragQueryFileW, (HDROP hDrop, UINT iFile, LPWSTR lpszFile, UINT cch))
NUI_FN(void, DragFinish, (HDROP hDrop))
NUI_FN(LPWSTR, GetCommandLineW, (void))
NUI_FN(LPWSTR *, CommandLineToArgvW, (LPCWSTR lpCmdLine, int *pNumArgs))
NUI_FN(HANDLE, LocalFree, (HANDLE hMem))
NUI_FN(BOOL, GetOpenFileNameW, (OPENFILENAMEW *lpofn))
NUI_FN(BOOL, GetSaveFileNameW, (OPENFILENAMEW *lpofn))

/* comctl32 — только через LoadLibraryW + GetProcAddress (ADR-16) */
NUI_FNPTR(BOOL, InitCommonControlsEx, (const INITCOMMONCONTROLSEX *picce))
NUI_FNPTR(HRESULT, TaskDialogIndirect, (const TASKDIALOGCONFIG *pTaskConfig, int *pnButton, int *pnRadioButton,
          BOOL *pfVerificationFlagChecked))
/* GDI+ (Image) — тоже через GetProcAddress: gdiplus.dll грузится при первой картинке */
NUI_FNPTR(GpStatus, GdiplusStartup, (ULONG_PTR *token, const GdiplusStartupInput *input, GdiplusStartupOutput *output))
NUI_FNPTR(GpStatus, GdipCreateBitmapFromStream, (IStream *stream, GpBitmap **bitmap))
NUI_FNPTR(GpStatus, GdipCreateBitmapFromScan0, (INT width, INT height, INT stride, PixelFormat format, BYTE *scan0,
          GpBitmap **bitmap))
NUI_FNPTR(GpStatus, GdipDisposeImage, (GpImage *image))
NUI_FNPTR(GpStatus, GdipGetImageWidth, (GpImage *image, UINT *width))
NUI_FNPTR(GpStatus, GdipGetImageHeight, (GpImage *image, UINT *height))
NUI_FNPTR(GpStatus, GdipImageRotateFlip, (GpImage *image, RotateFlipType rfType))
NUI_FNPTR(GpStatus, GdipGetImageGraphicsContext, (GpImage *image, GpGraphics **graphics))
NUI_FNPTR(GpStatus, GdipSetInterpolationMode, (GpGraphics *graphics, InterpolationMode interpolationMode))
NUI_FNPTR(GpStatus, GdipSetPixelOffsetMode, (GpGraphics *graphics, PixelOffsetMode pixelOffsetMode))
NUI_FNPTR(GpStatus, GdipDrawImageRectRectI, (GpGraphics *graphics, GpImage *image, INT dstx, INT dsty, INT dstwidth,
          INT dstheight, INT srcx, INT srcy, INT srcwidth, INT srcheight, GpUnit srcUnit,
          const GpImageAttributes *imageAttributes, DrawImageAbort callback, void *callbackData))
NUI_FNPTR(GpStatus, GdipDeleteGraphics, (GpGraphics *graphics))
NUI_FNPTR(GpStatus, GdipCreateImageAttributes, (GpImageAttributes **imageattr))
NUI_FNPTR(GpStatus, GdipSetImageAttributesWrapMode, (GpImageAttributes *imageAttr, WrapMode wrap, ARGB argb,
          BOOL clamp))
NUI_FNPTR(GpStatus, GdipDisposeImageAttributes, (GpImageAttributes *imageattr))
/* Canvas */
NUI_FNPTR(GpStatus, GdipCreateFromHDC, (HDC hdc, GpGraphics **graphics))
NUI_FNPTR(GpStatus, GdipSetSmoothingMode, (GpGraphics *graphics, SmoothingMode smoothingMode))
NUI_FNPTR(GpStatus, GdipScaleWorldTransform, (GpGraphics *graphics, REAL sx, REAL sy, GpMatrixOrder order))
NUI_FNPTR(GpStatus, GdipCreatePen1, (ARGB color, REAL width, GpUnit unit, GpPen **pen))
NUI_FNPTR(GpStatus, GdipDeletePen, (GpPen *pen))
NUI_FNPTR(GpStatus, GdipSetPenColor, (GpPen *pen, ARGB argb))
NUI_FNPTR(GpStatus, GdipSetPenWidth, (GpPen *pen, REAL width))
NUI_FNPTR(GpStatus, GdipCreateSolidFill, (ARGB color, GpSolidFill **brush))
NUI_FNPTR(GpStatus, GdipSetSolidFillColor, (GpSolidFill *brush, ARGB color))
NUI_FNPTR(GpStatus, GdipDeleteBrush, (GpBrush *brush))
NUI_FNPTR(GpStatus, GdipDrawLine, (GpGraphics *graphics, GpPen *pen, REAL x1, REAL y1, REAL x2, REAL y2))
NUI_FNPTR(GpStatus, GdipDrawRectangle, (GpGraphics *graphics, GpPen *pen, REAL x, REAL y, REAL width, REAL height))
NUI_FNPTR(GpStatus, GdipFillRectangle, (GpGraphics *graphics, GpBrush *brush, REAL x, REAL y, REAL width, REAL height))
NUI_FNPTR(GpStatus, GdipDrawEllipse, (GpGraphics *graphics, GpPen *pen, REAL x, REAL y, REAL width, REAL height))
NUI_FNPTR(GpStatus, GdipFillEllipse, (GpGraphics *graphics, GpBrush *brush, REAL x, REAL y, REAL width, REAL height))

/* ---------- 5. Обёртки (wrap_windows.c) ---------- */
/* Свои сообщения: WM_APP + n. */
enum { NUI_WM_DO = 0x8000 + 1, NUI_WM_RELAYOUT, NUI_WM_ENTER, NUI_WM_LISTSEL, NUI_WM_KEY, NUI_WM_NAVKEY };

int   nui_init(void);
HWND  nui_msg_window(void);
HWND  nui_parking(void);
void  nui_loop(void);
HWND  nui_window_new(LPCWSTR title, DWORD style, LONG_PTR node);
HWND  nui_control(LPCWSTR cls, LPCWSTR text, DWORD style, DWORD exstyle, HWND parent, LONG_PTR node);
HWND  nui_scroll_new(HWND parent, LONG_PTR node);
void  nui_fonts(UINT dpi, HFONT out[3]);
void  nui_metrics(HWND hwnd, HFONT font, int out[2]);
void  nui_text_size(HWND hwnd, HFONT font, LPCWSTR text, int len, int wrapWidth, int out[2]);
void  nui_ideal_size(HWND button, int out[2]);
void  nui_window_frame(HWND hwnd, int cw, int ch, int center, int out[4]);
void  nui_client_size(HWND hwnd, int out[2]);
void  nui_monitor_rect(HWND hwnd, int out[4]);
HFONT nui_font_scaled(HFONT base, int pct, int bold);
void  nui_fill(HDC dc, int x, int y, int w, int h, COLORREF c);
void  nui_draw_button(const DRAWITEMSTRUCT *d, unsigned fg, unsigned bg, int isDefault); /* цвета — COLORREF; ~0u — темы */
int   nui_message(HWND parent, int kind, LPCWSTR app, LPCWSTR title, LPCWSTR text);
int   nui_file_dialog(HWND parent, int mode, LPCWSTR title, LPCWSTR filter, LPWSTR buf, int cap);
int   nui_folder_dialog(HWND parent, LPCWSTR title, LPWSTR buf);
void  nui_text_append(HWND edit, LPCWSTR s, int autoscroll, int maxLines);
int   nui_clip_set(LPCWSTR s, int len);
int   nui_clip_get(LPWSTR buf, int cap); /* длина текста в буфере; buf == 0 — только длина */
void *nui_image_decode(const BYTE *data, int len, int orient, int wh[2]); /* 0 — не картинка или нет GDI+ */
HBITMAP nui_image_render(void *pic, int w, int h);           /* DIB 32 бита, альфа умножена */
void  nui_image_free(void *pic);
void  nui_draw_image(const DRAWITEMSTRUCT *d, HBITMAP bmp, int w, int h, HBRUSH bg);
void  nui_list_column(HWND lv, int index, LPCWSTR title);
HCURSOR nui_cursor(int id); /* системный курсор по номеру (IDC_*) */
void  nui_ctx_forward(HWND edit, int on); /* EDIT: WM_CONTEXTMENU — родителю, а не своё меню */
/* Canvas: рисование одного WM_PAINT (двойная буферизация); координаты — логические пиксели */
typedef struct NuiPaint NuiPaint;
NuiPaint *nui_canvas_begin(HWND h, HBRUSH bg, HFONT font, UINT dpi, COLORREF text);
void  nui_canvas_end(HWND h);
void  nui_paint_color(NuiPaint *p, COLORREF c);
void  nui_paint_width(NuiPaint *p, float w);
void  nui_paint_line(NuiPaint *p, float x1, float y1, float x2, float y2);
void  nui_paint_rect(NuiPaint *p, float x, float y, float w, float h, int fill);
void  nui_paint_ellipse(NuiPaint *p, float x, float y, float w, float h, int fill);
void  nui_paint_text(NuiPaint *p, float x, float y, LPCWSTR s, int len);
void  nui_paint_text_size(NuiPaint *p, LPCWSTR s, int len, float out[2]);
int   nui_menu_set_text(HMENU m, UINT item, int byPos, LPCWSTR text);
void  nui_edit_hint(HWND edit, LPCWSTR hint); /* подсказка в пустом многострочном поле; 0 — убрать */
void  nui_list_select(HWND lv, int row);
/* Tabs: SysTabControl32 */
void  nui_tabs_insert(HWND tab, int i, LPCWSTR title);
void  nui_tabs_inner(HWND tab, int w, int h, int out[4]); /* область страниц в окне w×h: x, y, ширина, высота */
void  nui_tabs_hole(HWND tab, int w, int h, const int in[4]); /* окно контрола без области страниц */
int   nui_tabs_item_width(HWND tab, int i);
COLORREF nui_tabs_pane(void); /* цвет области страниц у темы */
void  nui_tabs_click(HWND tab, int i);      /* только с NUI_GUI_TEST: щелчок по ярлыку i */
void  nui_sim_spin(HWND edit, int up);      /* только с NUI_GUI_TEST: ↑ или ↓ в поле Spin */
HWND  nui_hwnd_at(HWND win, int x, int y);  /* только с NUI_GUI_TEST: окно под точкой клиентской области */
int   nui_list_item_text(HWND lv, int row, int col, LPWSTR buf, int cap); /* только с NUI_GUI_TEST */
void  nui_sim_drop(HWND w, LPCWSTR files, int len); /* только с NUI_GUI_TEST: files — «путь\0путь\0\0» */
HWND  nui_tooltip_new(HWND owner);
void  nui_tooltip_set(HWND tip, HWND ctl, LPCWSTR text); /* text == 0 — убрать */
LRESULT nuiWndProc(HWND hwnd, UINT msg, WPARAM wp, LPARAM lp); /* //export в export_windows.go */
/* Только с NUI_GUI_TEST (sim_windows.go): открытый диалог и ответ за пользователя. */
int   nui_dialog_kind(void); /* 0 — нет; 1 — сообщение, 2 — вопрос, 3 — файл, 4 — папка */
HWND  nui_find_dialog(void);
int   nui_task_dialog(void); /* 1 — сообщения показывает TaskDialogIndirect */

#endif
