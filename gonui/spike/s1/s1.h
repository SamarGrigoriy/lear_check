/* s1.h — собственные прототипы GTK3/GLib для спайка S1 (ADR-3).
   Только объявления: файл подключается в преамбуле файла с //export.
   Режим NUI_CHECK (check-protos.sh) сверяет всё с настоящими заголовками. */
#ifndef NUI_S1_H
#define NUI_S1_H

/* ---------- 1. Базовые типы (в режиме проверки берутся из настоящих заголовков) ---------- */
#ifndef NUI_CHECK
typedef __UINTPTR_TYPE__ uintptr_t;
typedef __SIZE_TYPE__ size_t;
typedef int gint, gboolean;
typedef unsigned int guint;
typedef unsigned long gulong;
typedef char gchar;
typedef void *gpointer;
/* Перечисления: в ABI это int. В режиме проверки — настоящие enum. */
typedef int GtkWindowType, GtkWindowPosition, GtkOrientation, GtkDialogFlags,
            GtkMessageType, GtkButtonsType, GConnectFlags, GIOCondition;
typedef void (*GCallback)(void);
typedef gboolean (*GSourceFunc)(gpointer);
typedef gboolean (*GUnixFDSourceFunc)(gint, GIOCondition, gpointer);
typedef void (*GDestroyNotify)(gpointer);
typedef void (*GClosureNotify)(gpointer, void *);
/* Непрозрачные объекты */
typedef struct _GtkWidget GtkWidget;       typedef struct _GtkWindow GtkWindow;
typedef struct _GtkContainer GtkContainer; typedef struct _GtkBox GtkBox;
typedef struct _GtkLabel GtkLabel;         typedef struct _GtkButton GtkButton;
typedef struct _GtkDialog GtkDialog;
#endif

/* ---------- 2. Макросы ---------- */
#ifdef NUI_CHECK
#define NUI_FN(ret, name, params) \
	_Static_assert(__builtin_types_compatible_p(__typeof__(&name), ret (*) params), #name);
#define NUI_CONST(name, val) \
	_Static_assert((val) == (name), #name); enum { NUI_##name = (val) };
#else
#define NUI_FN(ret, name, params) ret name params;
#define NUI_CONST(name, val) enum { NUI_##name = (val) };
#endif

/* ---------- 3. Константы ---------- */
NUI_CONST(GTK_WINDOW_TOPLEVEL, 0)
NUI_CONST(GTK_WIN_POS_CENTER, 1)
NUI_CONST(GTK_ORIENTATION_VERTICAL, 1)
NUI_CONST(GTK_DIALOG_MODAL, 1)
NUI_CONST(GTK_DIALOG_DESTROY_WITH_PARENT, 2)
NUI_CONST(GTK_MESSAGE_QUESTION, 2)
NUI_CONST(GTK_BUTTONS_YES_NO, 4)
NUI_CONST(GTK_RESPONSE_YES, -8)
NUI_CONST(GTK_RESPONSE_NO, -9)
NUI_CONST(G_PRIORITY_DEFAULT, 0)
NUI_CONST(G_IO_IN, 1)
NUI_CONST(G_IO_ERR, 8)
NUI_CONST(G_IO_HUP, 16)
NUI_CONST(G_IO_NVAL, 32)

/* ---------- 4. Прототипы ---------- */
/* GLib, GObject, цикл событий */
NUI_FN(gboolean, gtk_init_check, (int *argc, char ***argv))
NUI_FN(void, gtk_main, (void))
NUI_FN(void, gtk_main_quit, (void))
NUI_FN(gulong, g_signal_connect_data, (gpointer instance, const gchar *detailed_signal,
       GCallback c_handler, gpointer data, GClosureNotify destroy_data, GConnectFlags connect_flags))
NUI_FN(guint, g_timeout_add_full, (gint priority, guint interval, GSourceFunc function,
       gpointer data, GDestroyNotify notify))
NUI_FN(guint, g_idle_add_full, (gint priority, GSourceFunc function, gpointer data, GDestroyNotify notify))
NUI_FN(guint, g_unix_fd_add_full, (gint priority, gint fd, GIOCondition condition,
       GUnixFDSourceFunc function, gpointer user_data, GDestroyNotify notify))
NUI_FN(gboolean, g_source_remove, (guint tag))
NUI_FN(size_t, strlen, (const char *s))
NUI_FN(int, gettid, (void))
NUI_FN(int, getpid, (void))

/* Окно, контейнеры, виджеты */
NUI_FN(GtkWidget *, gtk_window_new, (GtkWindowType type))
NUI_FN(void, gtk_window_set_title, (GtkWindow *window, const gchar *title))
NUI_FN(const gchar *, gtk_window_get_title, (GtkWindow *window))
NUI_FN(void, gtk_window_set_default_size, (GtkWindow *window, gint width, gint height))
NUI_FN(void, gtk_window_set_position, (GtkWindow *window, GtkWindowPosition position))
NUI_FN(void, gtk_window_present, (GtkWindow *window))
NUI_FN(void, gtk_window_close, (GtkWindow *window))
NUI_FN(void, gtk_container_add, (GtkContainer *container, GtkWidget *widget))
NUI_FN(void, gtk_container_set_border_width, (GtkContainer *container, guint border_width))
NUI_FN(GtkWidget *, gtk_box_new, (GtkOrientation orientation, gint spacing))
NUI_FN(void, gtk_box_pack_start, (GtkBox *box, GtkWidget *child, gboolean expand, gboolean fill, guint padding))
NUI_FN(GtkWidget *, gtk_label_new, (const gchar *str))
NUI_FN(void, gtk_label_set_text, (GtkLabel *label, const gchar *str))
NUI_FN(const gchar *, gtk_label_get_text, (GtkLabel *label))
NUI_FN(GtkWidget *, gtk_button_new_with_label, (const gchar *label))
NUI_FN(void, gtk_button_clicked, (GtkButton *button))
NUI_FN(void, gtk_widget_show, (GtkWidget *widget))
NUI_FN(void, gtk_widget_destroy, (GtkWidget *widget))

/* Диалог: varargs — только из обёрток */
NUI_FN(GtkWidget *, gtk_message_dialog_new, (GtkWindow *parent, GtkDialogFlags flags, GtkMessageType type,
       GtkButtonsType buttons, const gchar *message_format, ...))
NUI_FN(gint, gtk_dialog_run, (GtkDialog *dialog))
NUI_FN(void, gtk_dialog_response, (GtkDialog *dialog, gint response_id))

/* ---------- 5. Наши обёртки (wrap.c) и трамплины (export.go) ----------
   В обоих режимах: в режиме проверки wrap.c компилируется против настоящих заголовков. */
gulong nui_connect(gpointer obj, const gchar *sig, uintptr_t ud);
gulong nui_connect_ret(gpointer obj, const gchar *sig, uintptr_t ud);
guint  nui_timeout(guint ms, uintptr_t ud);
guint  nui_idle(uintptr_t ud);
guint  nui_watch(int fd, unsigned cond, uintptr_t ud);
int    nui_ask(GtkWindow *parent, const gchar *text);
void   nui_test_respond(int yes);

void nuiSignal(void *obj, uintptr_t ud);
int  nuiSignalRet(void *obj, void *arg, uintptr_t ud);
int  nuiSource(uintptr_t ud);
int  nuiWatch(int fd, unsigned cond, uintptr_t ud);

#endif
