/* gtk3_linux.h — собственные прототипы GTK3/GLib (ADR-3). Только объявления:
   файл подключается в преамбуле файла с //export.
   В режиме NUI_CHECK (tools/check-protos.sh) всё сверяется с настоящими заголовками. */
#ifndef NUI_GTK3_H
#define NUI_GTK3_H

/* ---------- 1. Базовые типы (в режиме проверки берутся из настоящих заголовков) ---------- */
#ifndef NUI_CHECK
typedef __UINTPTR_TYPE__ uintptr_t;
typedef __SIZE_TYPE__ size_t;
typedef int gint, gboolean;
typedef unsigned int guint;
typedef unsigned long gulong;
typedef char gchar;
typedef unsigned char guchar;
typedef unsigned long gsize;
typedef float gfloat;
typedef void *gpointer;
/* Перечисления: в ABI это int. В режиме проверки — настоящие enum. */
typedef int GtkWindowType, GtkWindowPosition, GtkOrientation, GtkAlign, GConnectFlags, GIOCondition,
            GtkPolicyType, GtkWrapMode, GtkJustification, PangoEllipsizeMode, PangoWeight,
            GtkDialogFlags, GtkMessageType, GtkButtonsType, GtkFileChooserAction, GBusType, GDBusCallFlags,
            GdkInterpType, GtkTreeModelFlags, GtkTreeViewColumnSizing, GtkSelectionMode, GTypeFlags,
            GtkStateFlags, GtkTextWindowType, PangoWrapMode, GtkDestDefaults, GdkDragAction,
            GdkModifierType, GdkEventType, GdkScrollDirection, GdkEventMask, GdkGravity, GtkSpinType;
typedef unsigned int guint32;
typedef unsigned long GType;
typedef unsigned short guint16;
typedef unsigned char guint8;
typedef double gdouble;
typedef long gssize;
typedef void (*GCallback)(void);
typedef gboolean (*GSourceFunc)(gpointer);
typedef gboolean (*GUnixFDSourceFunc)(gint, GIOCondition, gpointer);
typedef void (*GDestroyNotify)(gpointer);
typedef void (*GClosureNotify)(gpointer, void *);
/* Непрозрачные объекты */
typedef struct _GtkWidget GtkWidget;       typedef struct _GtkWindow GtkWindow;
typedef struct _GtkContainer GtkContainer; typedef struct _GtkBin GtkBin;
typedef struct _GtkBox GtkBox;             typedef struct _GtkLabel GtkLabel;
typedef struct _GtkButton GtkButton;       typedef struct _GtkToggleButton GtkToggleButton;
typedef struct _GtkEntry GtkEntry;         typedef struct _GtkEditable GtkEditable;
typedef struct _GtkComboBox GtkComboBox;   typedef struct _GtkComboBoxText GtkComboBoxText;
typedef struct _GtkRange GtkRange;         typedef struct _GtkScale GtkScale;
typedef struct _GtkProgressBar GtkProgressBar;
typedef struct _GtkGrid GtkGrid;           typedef struct _GtkFrame GtkFrame;
typedef struct _GtkScrolledWindow GtkScrolledWindow; typedef struct _GtkAdjustment GtkAdjustment;
typedef struct _GtkTextView GtkTextView;   typedef struct _GtkTextBuffer GtkTextBuffer;
typedef struct _GtkTextIter GtkTextIter;   typedef struct _GtkTextMark GtkTextMark;
typedef struct _PangoAttrList PangoAttrList; typedef struct _PangoAttribute PangoAttribute;
typedef struct _GObject GObject;           typedef struct _GTypeInstance GTypeInstance;
typedef struct _GtkDialog GtkDialog;       typedef struct _GtkMessageDialog GtkMessageDialog;
typedef struct _GtkFileChooser GtkFileChooser; typedef struct _GtkFileFilter GtkFileFilter;
typedef struct _GtkNativeDialog GtkNativeDialog;
typedef struct _GtkFileChooserNative GtkFileChooserNative;
typedef struct _GError GError;             typedef struct _GVariant GVariant;
typedef struct _GVariantType GVariantType; typedef struct _GDBusConnection GDBusConnection;
typedef struct _GCancellable GCancellable; typedef struct _GAsyncResult GAsyncResult;
typedef struct _GdkWindow GdkWindow;
typedef struct _GdkDisplay GdkDisplay;     typedef struct _GdkSeat GdkSeat;
typedef struct _GdkDevice GdkDevice;       typedef struct _GdkKeymap GdkKeymap;
typedef union _GdkEvent GdkEvent;
typedef struct _cairo cairo_t;
typedef struct _GdkRGBA { double red, green, blue, alpha; } GdkRGBA;
typedef struct _GdkRectangle GdkRectangle;
typedef struct _GdkKeymapKey { guint keycode; gint group, level; } GdkKeymapKey;
/* События мыши (Canvas): драйвер читает их функциями доступа GDK, тесты — заполняют. */
typedef struct _GdkEventButton {
	int type;
	GdkWindow *window;
	signed char send_event;
	guint32 time;
	double x, y, *axes;
	guint state, button;
	GdkDevice *device;
	double x_root, y_root;
} GdkEventButton;
typedef struct _GdkEventMotion {
	int type;
	GdkWindow *window;
	signed char send_event;
	guint32 time;
	double x, y, *axes;
	guint state;
	short is_hint;
	GdkDevice *device;
	double x_root, y_root;
} GdkEventMotion;
typedef struct _GdkEventScroll {
	int type;
	GdkWindow *window;
	signed char send_event;
	guint32 time;
	double x, y;
	guint state;
	int direction;
	GdkDevice *device;
	double x_root, y_root, delta_x, delta_y;
	guint is_stop; /* в GTK — битовое поле в этом же слове; cgo TinyGo битовых полей не понимает */
} GdkEventScroll;
/* GdkEventKey: драйвер читает keyval, state и hardware_keycode (OnKey); тесты — заполняют. */
typedef struct _GdkEventKey {
	int type;
	GdkWindow *window;
	signed char send_event;
	guint32 time;
	guint state, keyval;
	gint length;
	gchar *string;
	guint16 hardware_keycode;
	guint8 group;
	guint8 is_modifier; /* в GTK — битовое поле в этом же байте; cgo TinyGo битовых полей не понимает */
} GdkEventKey;
typedef struct _GtkCssProvider GtkCssProvider; typedef struct _GtkStyleContext GtkStyleContext;
typedef struct _GtkStyleProvider GtkStyleProvider;
typedef struct _GtkClipboard GtkClipboard; typedef struct _GdkAtom *GdkAtom;
typedef struct _GtkTargetEntry { gchar *target; guint flags, info; } GtkTargetEntry;
typedef struct _GdkDragContext GdkDragContext; typedef struct _GtkSelectionData GtkSelectionData;
typedef struct _GtkImage GtkImage;         typedef struct _GdkPixbuf GdkPixbuf;
typedef struct _GdkPixbufLoader GdkPixbufLoader; typedef struct _cairo_surface cairo_surface_t;
typedef void (*GAsyncReadyCallback)(GObject *, GAsyncResult *, gpointer);
typedef struct _GList GList;
struct _GList { gpointer data; GList *next; GList *prev; };
/* Одним объявлением: cgo TinyGo не видит полей структуры, описанной после typedef. */
typedef struct _GSList { gpointer data; struct _GSList *next; } GSList;
typedef struct _GtkEventBox GtkEventBox; typedef struct _GtkPaned GtkPaned; typedef struct _GtkMenu GtkMenu;
typedef struct _GtkNotebook GtkNotebook; typedef struct _GtkRadioButton GtkRadioButton;
typedef struct _GtkOrientable GtkOrientable; typedef struct _GtkSpinButton GtkSpinButton;
typedef void (*GtkMenuDetachFunc)(GtkWidget *attach_widget, GtkMenu *menu);
typedef struct _GtkMenuShell GtkMenuShell; typedef struct _GtkMenuItem GtkMenuItem;
typedef struct _GtkCheckMenuItem GtkCheckMenuItem; typedef struct _GtkAccelLabel GtkAccelLabel;
typedef struct _GtkTreeView GtkTreeView;   typedef struct _GtkTreeSelection GtkTreeSelection;
typedef struct _GtkTreePath GtkTreePath;   typedef struct _GtkTreeModel GtkTreeModel;
typedef void (*GtkCallback)(GtkWidget *, gpointer);
/* List: своя модель GtkTreeModel — нужны раскладки GObject, класса и интерфейса.
   Размеры и смещения сверяются с настоящими (раздел 3). */
typedef struct _GtkTreeViewColumn GtkTreeViewColumn; typedef struct _GtkCellRenderer GtkCellRenderer;
typedef struct _GValue GValue;             typedef struct _PangoLayout PangoLayout;
struct _GObject { gpointer g_type_instance; guint ref_count; gpointer qdata; };
typedef struct _GObjectClass {
	GType g_type; /* GTypeClass */
	gpointer construct_properties, constructor, set_property, get_property, dispose;
	void (*finalize)(GObject *object);
	gpointer rest[10]; /* dispatch_properties_changed … pdummy */
} GObjectClass;
typedef struct _GtkTreeIter { gint stamp; gpointer user_data, user_data2, user_data3; } GtkTreeIter;
typedef struct { GType g_type, g_instance_type; } GTypeInterface;
typedef struct _GtkTreeModelIface {
	GTypeInterface g_iface;
	gpointer row_changed, row_inserted, row_has_child_toggled, row_deleted, rows_reordered; /* сигналы */
	GtkTreeModelFlags (*get_flags)(GtkTreeModel *tree_model);
	gint (*get_n_columns)(GtkTreeModel *tree_model);
	GType (*get_column_type)(GtkTreeModel *tree_model, gint index_);
	gboolean (*get_iter)(GtkTreeModel *tree_model, GtkTreeIter *iter, GtkTreePath *path);
	GtkTreePath *(*get_path)(GtkTreeModel *tree_model, GtkTreeIter *iter);
	void (*get_value)(GtkTreeModel *tree_model, GtkTreeIter *iter, gint column, GValue *value);
	gboolean (*iter_next)(GtkTreeModel *tree_model, GtkTreeIter *iter);
	gboolean (*iter_previous)(GtkTreeModel *tree_model, GtkTreeIter *iter);
	gboolean (*iter_children)(GtkTreeModel *tree_model, GtkTreeIter *iter, GtkTreeIter *parent);
	gboolean (*iter_has_child)(GtkTreeModel *tree_model, GtkTreeIter *iter);
	gint (*iter_n_children)(GtkTreeModel *tree_model, GtkTreeIter *iter);
	gboolean (*iter_nth_child)(GtkTreeModel *tree_model, GtkTreeIter *iter, GtkTreeIter *parent, gint n);
	gboolean (*iter_parent)(GtkTreeModel *tree_model, GtkTreeIter *iter, GtkTreeIter *child);
	void (*ref_node)(GtkTreeModel *tree_model, GtkTreeIter *iter);
	void (*unref_node)(GtkTreeModel *tree_model, GtkTreeIter *iter);
} GtkTreeModelIface;
typedef void (*GClassInitFunc)(gpointer g_class, gpointer class_data);
typedef void (*GInstanceInitFunc)(GTypeInstance *instance, gpointer g_class);
typedef void (*GInterfaceInitFunc)(gpointer g_iface, gpointer iface_data);
typedef struct { GInterfaceInitFunc interface_init; GDestroyNotify interface_finalize; gpointer interface_data; } GInterfaceInfo;
/* GtkAllocation (= GdkRectangle) — только для тестового помощника nui_allocation */
typedef struct { int x, y, width, height; } GtkAllocation;
#endif

/* Место под GtkTextIter на стеке C: настоящий размер ~80 байт, берём с запасом.
   В режиме проверки размер сверяется (раздел 3). Go к итераторам не прикасается. */
typedef struct { void *p[16]; } NuiTextIter;

/* ---------- 2. Макросы ---------- */
#ifdef NUI_CHECK
#define NUI_FN(ret, name, params) \
	_Static_assert(__builtin_types_compatible_p(__typeof__(&name), ret (*) params), #name);
#define NUI_CONST(name, val) \
	_Static_assert((val) == (name), #name); enum { NUI_##name = (val) };
/* Функции без аргументов с G_GNUC_CONST (*_get_type): атрибут const у GCC — часть типа
   функции, поэтому сверяется тип результата. */
#define NUI_FN_CONST(ret, name) \
	_Static_assert(__builtin_types_compatible_p(__typeof__(name()), ret), #name);
#else
#define NUI_FN(ret, name, params) ret name params;
#define NUI_CONST(name, val) enum { NUI_##name = (val) };
#define NUI_FN_CONST(ret, name) ret name(void) __attribute__((const));
#endif

/* ---------- 3. Константы ---------- */
NUI_CONST(GTK_WINDOW_TOPLEVEL, 0)
NUI_CONST(GTK_WIN_POS_CENTER, 1)
NUI_CONST(GTK_ORIENTATION_HORIZONTAL, 0)
NUI_CONST(GTK_ORIENTATION_VERTICAL, 1)
NUI_CONST(GTK_SPIN_STEP_FORWARD, 0)
NUI_CONST(GTK_SPIN_STEP_BACKWARD, 1)
NUI_CONST(GTK_ALIGN_FILL, 0)
NUI_CONST(GTK_ALIGN_START, 1)
NUI_CONST(GTK_ALIGN_END, 2)
NUI_CONST(GTK_ALIGN_CENTER, 3)
NUI_CONST(G_PRIORITY_DEFAULT, 0)
NUI_CONST(G_IO_IN, 1)
NUI_CONST(G_IO_PRI, 2)
NUI_CONST(G_IO_OUT, 4)
NUI_CONST(G_IO_ERR, 8)
NUI_CONST(G_IO_HUP, 16)
NUI_CONST(G_IO_NVAL, 32)
NUI_CONST(GTK_POLICY_AUTOMATIC, 1)
NUI_CONST(GTK_POLICY_NEVER, 2)
NUI_CONST(GTK_WRAP_NONE, 0)
NUI_CONST(GTK_WRAP_WORD_CHAR, 3)
NUI_CONST(GTK_STYLE_PROVIDER_PRIORITY_APPLICATION, 600)
NUI_CONST(GTK_JUSTIFY_LEFT, 0)
NUI_CONST(GTK_JUSTIFY_RIGHT, 1)
NUI_CONST(GTK_JUSTIFY_CENTER, 2)
NUI_CONST(PANGO_ELLIPSIZE_NONE, 0)
NUI_CONST(PANGO_ELLIPSIZE_START, 1)
NUI_CONST(PANGO_ELLIPSIZE_MIDDLE, 2)
NUI_CONST(PANGO_ELLIPSIZE_END, 3)
NUI_CONST(PANGO_WEIGHT_NORMAL, 400)
NUI_CONST(PANGO_WEIGHT_BOLD, 700)
NUI_CONST(GTK_DIALOG_MODAL, 1)
NUI_CONST(GTK_DIALOG_DESTROY_WITH_PARENT, 2)
NUI_CONST(GTK_MESSAGE_INFO, 0)
NUI_CONST(GTK_MESSAGE_QUESTION, 2)
NUI_CONST(GTK_MESSAGE_ERROR, 3)
NUI_CONST(GTK_BUTTONS_OK, 1)
NUI_CONST(GTK_BUTTONS_YES_NO, 4)
NUI_CONST(GTK_RESPONSE_ACCEPT, -3)
NUI_CONST(GTK_RESPONSE_DELETE_EVENT, -4)
NUI_CONST(GTK_RESPONSE_OK, -5)
NUI_CONST(GTK_RESPONSE_CANCEL, -6)
NUI_CONST(GTK_RESPONSE_YES, -8)
NUI_CONST(GTK_RESPONSE_NO, -9)
NUI_CONST(GDK_INTERP_BILINEAR, 2)
NUI_CONST(GTK_FILE_CHOOSER_ACTION_OPEN, 0)
NUI_CONST(GTK_FILE_CHOOSER_ACTION_SAVE, 1)
NUI_CONST(GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER, 2)
NUI_CONST(G_BUS_TYPE_SESSION, 2)
NUI_CONST(G_DBUS_CALL_FLAGS_NONE, 0)
NUI_CONST(GDK_KEY_PRESS, 8)
NUI_CONST(GDK_MOTION_NOTIFY, 3)
NUI_CONST(GDK_BUTTON_PRESS, 4)
NUI_CONST(GDK_BUTTON_RELEASE, 7)
NUI_CONST(GDK_SCROLL, 31)
NUI_CONST(GDK_SCROLL_UP, 0)
NUI_CONST(GDK_SCROLL_DOWN, 1)
NUI_CONST(GDK_SCROLL_LEFT, 2)
NUI_CONST(GDK_SCROLL_RIGHT, 3)
NUI_CONST(GDK_POINTER_MOTION_MASK, 4)
NUI_CONST(GDK_BUTTON_PRESS_MASK, 256)
NUI_CONST(GDK_BUTTON_RELEASE_MASK, 512)
NUI_CONST(GDK_SCROLL_MASK, 2097152)
NUI_CONST(GTK_DEST_DEFAULT_ALL, 7)
NUI_CONST(GDK_ACTION_COPY, 2)
NUI_CONST(GTK_TEXT_WINDOW_WIDGET, 1)
NUI_CONST(PANGO_WRAP_WORD_CHAR, 2)
NUI_CONST(PANGO_SCALE, 1024)
NUI_CONST(GDK_SHIFT_MASK, 1)
NUI_CONST(GDK_CONTROL_MASK, 4)
NUI_CONST(GDK_MOD1_MASK, 8)
NUI_CONST(G_TYPE_STRING, 64)
NUI_CONST(G_TYPE_OBJECT, 80)
NUI_CONST(GTK_TREE_MODEL_ITERS_PERSIST, 1)
NUI_CONST(GTK_TREE_MODEL_LIST_ONLY, 2)
NUI_CONST(GTK_TREE_VIEW_COLUMN_FIXED, 2)
NUI_CONST(GTK_SELECTION_SINGLE, 1)
#ifdef NUI_CHECK
_Static_assert(sizeof(NuiTextIter) >= sizeof(GtkTextIter), "NuiTextIter меньше GtkTextIter");
_Static_assert(sizeof(GType) == sizeof(unsigned long), "GType");
_Static_assert(sizeof(GdkRectangle) == 4 * sizeof(int), "GdkRectangle — int[4] в text_draw_hint");
#endif
_Static_assert(sizeof(GdkRGBA) == 32, "GdkRGBA");
/* Раскладки для модели List: те же числа в обоих режимах (x86-64, arm64). */
_Static_assert(sizeof(GObject) == 24, "GObject");
_Static_assert(sizeof(GObjectClass) == 136, "GObjectClass");
_Static_assert(__builtin_offsetof(GObjectClass, finalize) == 48, "GObjectClass.finalize");
_Static_assert(sizeof(GtkTreeIter) == 32, "GtkTreeIter");
_Static_assert(__builtin_offsetof(GtkTreeModelIface, get_flags) == 56, "GtkTreeModelIface.get_flags");
_Static_assert(sizeof(GtkTreeModelIface) == 176, "GtkTreeModelIface");
_Static_assert(sizeof(GInterfaceInfo) == 24, "GInterfaceInfo");
_Static_assert(sizeof(GdkEventKey) == 56, "GdkEventKey");
_Static_assert(__builtin_offsetof(GdkEventKey, keyval) == 28, "GdkEventKey.keyval");
_Static_assert(__builtin_offsetof(GdkEventKey, hardware_keycode) == 48, "GdkEventKey.hardware_keycode");
_Static_assert(sizeof(GdkKeymapKey) == 12, "GdkKeymapKey");
_Static_assert(sizeof(GtkTargetEntry) == 16, "GtkTargetEntry");
_Static_assert(sizeof(GdkEventButton) == 80 && __builtin_offsetof(GdkEventButton, button) == 52, "GdkEventButton");
_Static_assert(sizeof(GdkEventMotion) == 80 && __builtin_offsetof(GdkEventMotion, device) == 56, "GdkEventMotion");
_Static_assert(sizeof(GdkEventScroll) == 96 && __builtin_offsetof(GdkEventScroll, direction) == 44, "GdkEventScroll");

/* ---------- 4. Прототипы ---------- */
/* GLib, GObject, цикл событий. (w) — только из обёрток. */
NUI_FN(gboolean, gtk_init_check, (int *argc, char ***argv))
NUI_FN(void, gtk_main, (void))
NUI_FN(void, gtk_main_quit, (void))
NUI_FN(void, g_set_prgname, (const gchar *prgname))
NUI_FN(gulong, g_signal_connect_data, (gpointer instance, const gchar *detailed_signal,
       GCallback c_handler, gpointer data, GClosureNotify destroy_data, GConnectFlags connect_flags)) /* (w) */
NUI_FN(guint, g_timeout_add_full, (gint priority, guint interval, GSourceFunc function,
       gpointer data, GDestroyNotify notify))                                                       /* (w) */
NUI_FN(guint, g_idle_add_full, (gint priority, GSourceFunc function, gpointer data, GDestroyNotify notify)) /* (w) */
NUI_FN(guint, g_unix_fd_add_full, (gint priority, gint fd, GIOCondition condition,
       GUnixFDSourceFunc function, gpointer user_data, GDestroyNotify notify))                      /* (w) */
NUI_FN(gboolean, g_source_remove, (guint tag))
NUI_FN(void, g_free, (gpointer mem))
NUI_FN(void, g_object_set_data, (GObject *object, const gchar *key, gpointer data))   /* (w) */
NUI_FN(gpointer, g_object_get_data, (GObject *object, const gchar *key))              /* (w) */
NUI_FN(gpointer, g_object_ref, (gpointer object))                                      /* (w) */
NUI_FN(void, g_object_unref, (gpointer object))
NUI_FN(void, g_object_set_data_full, (GObject *object, const gchar *key, gpointer data, GDestroyNotify destroy)) /* (w) */
NUI_FN(gchar *, g_strdup, (const gchar *str))                                          /* (w) */
NUI_FN(void, g_error_free, (GError *error))
NUI_FN(gchar *, g_filename_to_uri, (const gchar *filename, const gchar *hostname, GError **error)) /* g_free */
NUI_FN(gboolean, g_type_check_instance_is_a, (GTypeInstance *instance, GType iface_type))  /* тесты */
NUI_FN(void, g_list_free, (GList *list))
NUI_FN(void, g_slist_free, (GSList *list))
NUI_FN(size_t, strlen, (const char *s))
NUI_FN(int, gettid, (void)) /* pid_t; glibc ≥ 2.30 — ниже фактического минимума 2.34 */
NUI_FN(int, getpid, (void))

/* GtkWidget, GtkContainer, GtkBin */
NUI_FN(void, gtk_widget_show, (GtkWidget *widget))
NUI_FN(void, gtk_widget_hide, (GtkWidget *widget))
NUI_FN(void, gtk_widget_destroy, (GtkWidget *widget))
NUI_FN(void, gtk_widget_set_sensitive, (GtkWidget *widget, gboolean sensitive))
NUI_FN(gboolean, gtk_widget_get_sensitive, (GtkWidget *widget))
NUI_FN(void, gtk_widget_set_visible, (GtkWidget *widget, gboolean visible))
NUI_FN(gboolean, gtk_widget_get_visible, (GtkWidget *widget))
NUI_FN(void, gtk_widget_set_tooltip_text, (GtkWidget *widget, const gchar *text))
NUI_FN(void, gtk_widget_set_halign, (GtkWidget *widget, GtkAlign align))
NUI_FN(void, gtk_widget_set_valign, (GtkWidget *widget, GtkAlign align))
NUI_FN(void, gtk_widget_set_hexpand, (GtkWidget *widget, gboolean expand))
NUI_FN(void, gtk_widget_set_vexpand, (GtkWidget *widget, gboolean expand))
NUI_FN(void, gtk_widget_set_size_request, (GtkWidget *widget, gint width, gint height))
NUI_FN(void, gtk_widget_grab_focus, (GtkWidget *widget))
NUI_FN(void, gtk_widget_get_allocation, (GtkWidget *widget, GtkAllocation *allocation)) /* (w), тесты */
NUI_FN(void, gtk_widget_get_preferred_width, (GtkWidget *widget, gint *minimum_width, gint *natural_width))     /* тесты */
NUI_FN(void, gtk_widget_get_preferred_height, (GtkWidget *widget, gint *minimum_height, gint *natural_height))
NUI_FN(gboolean, gtk_widget_translate_coordinates, (GtkWidget *src_widget, GtkWidget *dest_widget,
       gint src_x, gint src_y, gint *dest_x, gint *dest_y))
NUI_FN(GtkWidget *, gtk_widget_get_toplevel, (GtkWidget *widget))
NUI_FN(GdkWindow *, gtk_widget_get_window, (GtkWidget *widget))                                                /* тесты */
NUI_FN(gint, gdk_window_get_origin, (GdkWindow *window, gint *x, gint *y))
NUI_FN(void, gtk_container_add, (GtkContainer *container, GtkWidget *widget))
NUI_FN(void, gtk_container_remove, (GtkContainer *container, GtkWidget *widget))
NUI_FN(void, gtk_container_set_border_width, (GtkContainer *container, guint border_width))
NUI_FN(guint, gtk_container_get_border_width, (GtkContainer *container))                     /* (w) */
NUI_FN(GtkStyleContext *, gtk_widget_get_style_context, (GtkWidget *widget))                  /* (w) */
NUI_FN(GtkClipboard *, gtk_clipboard_get, (GdkAtom selection))                              /* (w) */
NUI_FN(void, gtk_clipboard_set_text, (GtkClipboard *clipboard, const gchar *text, gint len))
NUI_FN(void, gtk_clipboard_set_can_store, (GtkClipboard *clipboard, const GtkTargetEntry *targets, gint n_targets))
NUI_FN(gchar *, gtk_clipboard_wait_for_text, (GtkClipboard *clipboard)) /* g_free */
/* Картинка (Image): GdkPixbuf декодирует, cairo-поверхность с масштабом экрана — чёткость на HiDPI */
NUI_FN(GtkWidget *, gtk_image_new, (void))                                                   /* (w) */
NUI_FN(void, gtk_image_set_from_surface, (GtkImage *image, cairo_surface_t *surface))       /* (w) */
NUI_FN(void, gtk_image_clear, (GtkImage *image))                                             /* (w) */
NUI_FN(gint, gtk_widget_get_scale_factor, (GtkWidget *widget))                               /* (w) */
NUI_FN(GdkPixbufLoader *, gdk_pixbuf_loader_new, (void))                                     /* (w) */
NUI_FN(gboolean, gdk_pixbuf_loader_write, (GdkPixbufLoader *loader, const guchar *buf, gsize count, GError **error)) /* (w) */
NUI_FN(gboolean, gdk_pixbuf_loader_close, (GdkPixbufLoader *loader, GError **error))        /* (w) */
NUI_FN(GdkPixbuf *, gdk_pixbuf_loader_get_pixbuf, (GdkPixbufLoader *loader))                 /* (w) */
NUI_FN(GdkPixbuf *, gdk_pixbuf_apply_embedded_orientation, (GdkPixbuf *src))                 /* (w) */
NUI_FN(int, gdk_pixbuf_get_width, (const GdkPixbuf *pixbuf))                                 /* (w) */
NUI_FN(int, gdk_pixbuf_get_height, (const GdkPixbuf *pixbuf))                                /* (w) */
NUI_FN(GdkPixbuf *, gdk_pixbuf_scale_simple, (const GdkPixbuf *src, int dest_width, int dest_height,
       GdkInterpType interp_type))                                                           /* (w) */
NUI_FN(cairo_surface_t *, gdk_cairo_surface_create_from_pixbuf, (const GdkPixbuf *pixbuf, int scale,
       GdkWindow *for_window))                                                               /* (w) */
NUI_FN(void, cairo_surface_destroy, (cairo_surface_t *surface))                              /* (w) */
NUI_FN(gpointer, g_malloc, (gsize n_bytes))                                                  /* (w) */
NUI_FN(gpointer, g_malloc0, (gsize n_bytes))                                                 /* (w) */
NUI_FN(void *, memcpy, (void *dest, const void *src, size_t n))                              /* (w) */
NUI_FN(gboolean, g_file_get_contents, (const gchar *filename, gchar **contents, gsize *length, GError **error))
/* List: GtkTreeView со своей моделью (строки спрашиваются у Go) */
NUI_FN(GType, g_type_register_static_simple, (GType parent_type, const gchar *type_name, guint class_size,
       GClassInitFunc class_init, guint instance_size, GInstanceInitFunc instance_init, GTypeFlags flags)) /* (w) */
NUI_FN(void, g_type_add_interface_static, (GType instance_type, GType interface_type, const GInterfaceInfo *info)) /* (w) */
NUI_FN(gpointer, g_type_class_peek_parent, (gpointer g_class))                              /* (w) */
NUI_FN(gpointer, g_object_new, (GType object_type, const gchar *first_property_name, ...))    /* (w) */
NUI_FN(void, g_object_set, (gpointer object, const gchar *first_property_name, ...))         /* (w) */
NUI_FN(GValue *, g_value_init, (GValue *value, GType g_type))                                 /* (w) */
NUI_FN(void, g_value_take_string, (GValue *value, gchar *v_string))                           /* (w) */
NUI_FN(gchar *, g_strndup, (const gchar *str, gsize n))                                       /* (w) */
NUI_FN_CONST(GType, gtk_tree_model_get_type)                                                  /* (w) */
NUI_FN(gint, gtk_tree_path_get_depth, (GtkTreePath *path))                                    /* (w) */
NUI_FN(gint *, gtk_tree_path_get_indices, (GtkTreePath *path))                                /* (w) */
NUI_FN(GtkTreePath *, gtk_tree_path_new_from_indices, (gint first_index, ...))                /* (w) */
NUI_FN(gboolean, gtk_tree_model_iter_nth_child, (GtkTreeModel *tree_model, GtkTreeIter *iter,
       GtkTreeIter *parent, gint n))                                                          /* (w) */
NUI_FN(void, gtk_tree_model_get, (GtkTreeModel *tree_model, GtkTreeIter *iter, ...))          /* тесты */
NUI_FN(GtkWidget *, gtk_tree_view_new, (void))                                                /* (w) */
NUI_FN(void, gtk_tree_view_set_model, (GtkTreeView *tree_view, GtkTreeModel *model))          /* (w) */
NUI_FN(GtkTreeModel *, gtk_tree_view_get_model, (GtkTreeView *tree_view))                     /* (w) */
NUI_FN(void, gtk_tree_view_set_fixed_height_mode, (GtkTreeView *tree_view, gboolean enable))  /* (w) */
NUI_FN(void, gtk_tree_view_set_enable_search, (GtkTreeView *tree_view, gboolean enable_search)) /* (w) */
NUI_FN(void, gtk_tree_view_set_headers_visible, (GtkTreeView *tree_view, gboolean headers_visible)) /* (w) */
NUI_FN(gint, gtk_tree_view_append_column, (GtkTreeView *tree_view, GtkTreeViewColumn *column)) /* (w) */
NUI_FN(guint, gtk_tree_view_get_n_columns, (GtkTreeView *tree_view))                          /* (w) */
NUI_FN(GtkTreeViewColumn *, gtk_tree_view_get_column, (GtkTreeView *tree_view, gint n))       /* (w) */
NUI_FN(void, gtk_tree_view_set_cursor, (GtkTreeView *tree_view, GtkTreePath *path, GtkTreeViewColumn *focus_column,
       gboolean start_editing))                                                               /* (w) */
NUI_FN(void, gtk_tree_view_row_activated, (GtkTreeView *tree_view, GtkTreePath *path, GtkTreeViewColumn *column)) /* тесты */
NUI_FN(GtkTreeViewColumn *, gtk_tree_view_column_new, (void))                                 /* (w) */
NUI_FN(void, gtk_tree_view_column_set_title, (GtkTreeViewColumn *tree_column, const gchar *title)) /* (w) */
NUI_FN(const gchar *, gtk_tree_view_column_get_title, (GtkTreeViewColumn *tree_column))        /* (w) */
NUI_FN(void, gtk_tree_view_column_pack_start, (GtkTreeViewColumn *tree_column, GtkCellRenderer *cell, gboolean expand)) /* (w) */
NUI_FN(void, gtk_tree_view_column_add_attribute, (GtkTreeViewColumn *tree_column, GtkCellRenderer *cell_renderer,
       const gchar *attribute, gint column))                                                  /* (w) */
NUI_FN(void, gtk_tree_view_column_set_sizing, (GtkTreeViewColumn *tree_column, GtkTreeViewColumnSizing type)) /* (w) */
NUI_FN(void, gtk_tree_view_column_set_fixed_width, (GtkTreeViewColumn *tree_column, gint fixed_width)) /* (w) */
NUI_FN(void, gtk_tree_view_column_set_resizable, (GtkTreeViewColumn *tree_column, gboolean resizable)) /* (w) */
NUI_FN(void, gtk_tree_view_column_set_expand, (GtkTreeViewColumn *tree_column, gboolean expand)) /* (w) */
NUI_FN(GtkCellRenderer *, gtk_cell_renderer_text_new, (void))                                 /* (w) */
NUI_FN(void, gtk_tree_selection_set_mode, (GtkTreeSelection *selection, GtkSelectionMode type)) /* (w) */
NUI_FN(gboolean, gtk_tree_selection_get_selected, (GtkTreeSelection *selection, GtkTreeModel **model, GtkTreeIter *iter)) /* (w) */
NUI_FN(void, gtk_tree_selection_unselect_all, (GtkTreeSelection *selection))                 /* (w) */
NUI_FN(PangoLayout *, gtk_widget_create_pango_layout, (GtkWidget *widget, const gchar *text)) /* (w) */
NUI_FN(void, pango_layout_set_text, (PangoLayout *layout, const char *text, int length))     /* (w) */
NUI_FN(void, pango_layout_get_pixel_size, (PangoLayout *layout, int *width, int *height))    /* (w) */
NUI_FN(void, gtk_widget_queue_draw, (GtkWidget *widget))
/* OnKey: клавиша по положению — через раскладки (GdkKeymap) */
NUI_FN(GdkDisplay *, gdk_display_get_default, (void))                                         /* (w) */
NUI_FN(GdkKeymap *, gdk_keymap_get_for_display, (GdkDisplay *display))                       /* (w) */
NUI_FN(gboolean, gdk_keymap_get_entries_for_keycode, (GdkKeymap *keymap, guint hardware_keycode,
       GdkKeymapKey **keys, guint **keyvals, gint *n_entries))                               /* (w) */
NUI_FN(gboolean, gdk_keymap_get_entries_for_keyval, (GdkKeymap *keymap, guint keyval,
       GdkKeymapKey **keys, gint *n_keys))                                                   /* тесты */
NUI_FN(GdkEvent *, gdk_event_new, (int type))                                                 /* тесты */
NUI_FN(void, gdk_event_free, (GdkEvent *event))                                               /* тесты */
NUI_FN(void, gdk_event_set_device, (GdkEvent *event, GdkDevice *device))                      /* тесты */
NUI_FN(GdkSeat *, gdk_display_get_default_seat, (GdkDisplay *display))                       /* тесты */
NUI_FN(GdkDevice *, gdk_seat_get_keyboard, (GdkSeat *seat))                                  /* тесты */
NUI_FN(GdkDevice *, gdk_seat_get_pointer, (GdkSeat *seat))                                   /* тесты */
NUI_FN(GdkPixbuf *, gdk_pixbuf_get_from_window, (GdkWindow *window, gint src_x, gint src_y, gint width, gint height)) /* тесты */
NUI_FN(guchar *, gdk_pixbuf_get_pixels, (const GdkPixbuf *pixbuf))                           /* тесты */
NUI_FN(gboolean, gtk_widget_event, (GtkWidget *widget, GdkEvent *event))                     /* тесты */
/* Контекстное меню */
NUI_FN(void, gtk_menu_attach_to_widget, (GtkMenu *menu, GtkWidget *attach_widget, GtkMenuDetachFunc detacher)) /* (w) */
NUI_FN_CONST(GType, gtk_scrolled_window_get_type)                                          /* (w) */
NUI_FN(void, gtk_menu_popup_at_pointer, (GtkMenu *menu, const GdkEvent *trigger_event))            /* (w) */
NUI_FN(void, gtk_menu_popup_at_widget, (GtkMenu *menu, GtkWidget *widget, GdkGravity widget_anchor,
       GdkGravity menu_anchor, const GdkEvent *trigger_event))                                    /* (w) */
NUI_FN(void, gtk_menu_popdown, (GtkMenu *menu))                                                    /* тесты */
NUI_FN(gboolean, gtk_widget_is_toplevel, (GtkWidget *widget))                                       /* (w) */
NUI_FN(gboolean, gtk_tree_view_get_path_at_pos, (GtkTreeView *tree_view, gint x, gint y, GtkTreePath **path,
       GtkTreeViewColumn **column, gint *cell_x, gint *cell_y))                                   /* (w) */
/* Split: GtkPaned */
NUI_FN(GtkWidget *, gtk_paned_new, (GtkOrientation orientation))
NUI_FN(void, gtk_paned_pack1, (GtkPaned *paned, GtkWidget *child, gboolean resize, gboolean shrink))
NUI_FN(void, gtk_paned_pack2, (GtkPaned *paned, GtkWidget *child, gboolean resize, gboolean shrink))
NUI_FN(void, gtk_paned_set_position, (GtkPaned *paned, gint position))
NUI_FN(gint, gtk_paned_get_position, (GtkPaned *paned))
/* Tabs: GtkNotebook */
NUI_FN(GtkWidget *, gtk_notebook_new, (void))                                               /* (w) */
NUI_FN(void, gtk_notebook_set_scrollable, (GtkNotebook *notebook, gboolean scrollable))    /* (w) */
NUI_FN(gint, gtk_notebook_append_page, (GtkNotebook *notebook, GtkWidget *child, GtkWidget *tab_label))
NUI_FN(void, gtk_notebook_set_current_page, (GtkNotebook *notebook, gint page_num))
NUI_FN(gint, gtk_notebook_get_current_page, (GtkNotebook *notebook))
NUI_FN(GtkWidget *, gtk_notebook_get_nth_page, (GtkNotebook *notebook, gint page_num))      /* (w) */
NUI_FN(void, gtk_notebook_set_tab_label_text, (GtkNotebook *notebook, GtkWidget *child, const gchar *tab_text))
/* Radio: GtkRadioButton в GtkBox */
NUI_FN(GtkWidget *, gtk_radio_button_new_with_label_from_widget, (GtkRadioButton *radio_group_member,
       const gchar *label))                                                                 /* (w) */
NUI_FN(void, gtk_orientable_set_orientation, (GtkOrientable *orientable, GtkOrientation orientation))
NUI_FN(GList *, g_list_nth, (GList *list, guint n))                                          /* (w) */
/* Spin: GtkSpinButton */
NUI_FN(GtkWidget *, gtk_spin_button_new_with_range, (gdouble min, gdouble max, gdouble step))
NUI_FN(void, gtk_spin_button_set_numeric, (GtkSpinButton *spin_button, gboolean numeric))
NUI_FN(void, gtk_spin_button_set_value, (GtkSpinButton *spin_button, gdouble value))
NUI_FN(gint, gtk_spin_button_get_value_as_int, (GtkSpinButton *spin_button))
NUI_FN(void, gtk_spin_button_spin, (GtkSpinButton *spin_button, GtkSpinType direction, gdouble increment)) /* тесты */
/* Canvas: GtkDrawingArea, cairo, события мыши */
NUI_FN(GtkWidget *, gtk_drawing_area_new, (void))
NUI_FN(void, gtk_widget_add_events, (GtkWidget *widget, gint events))
NUI_FN(void, cairo_set_source_rgb, (cairo_t *cr, double red, double green, double blue))     /* (w) */
NUI_FN(void, cairo_set_line_width, (cairo_t *cr, double width))                              /* (w) */
NUI_FN(void, cairo_line_to, (cairo_t *cr, double x, double y))                               /* (w) */
NUI_FN(void, cairo_stroke, (cairo_t *cr))                                                    /* (w) */
NUI_FN(void, cairo_fill, (cairo_t *cr))                                                      /* (w) */
NUI_FN(void, cairo_rectangle, (cairo_t *cr, double x, double y, double width, double height)) /* (w) */
NUI_FN(void, cairo_new_path, (cairo_t *cr))                                                  /* (w) */
NUI_FN(void, cairo_save, (cairo_t *cr))                                                      /* (w) */
NUI_FN(void, cairo_restore, (cairo_t *cr))                                                   /* (w) */
NUI_FN(void, cairo_translate, (cairo_t *cr, double tx, double ty))                           /* (w) */
NUI_FN(void, cairo_scale, (cairo_t *cr, double sx, double sy))                              /* (w) */
NUI_FN(void, cairo_arc, (cairo_t *cr, double xc, double yc, double radius, double angle1, double angle2)) /* (w) */
NUI_FN(GdkEventType, gdk_event_get_event_type, (const GdkEvent *event))                     /* (w) */
NUI_FN(gboolean, gdk_event_get_coords, (const GdkEvent *event, gdouble *x_win, gdouble *y_win)) /* (w) */
NUI_FN(gboolean, gdk_event_get_button, (const GdkEvent *event, guint *button))              /* (w) */
NUI_FN(gboolean, gdk_event_get_state, (const GdkEvent *event, GdkModifierType *state))      /* (w) */
NUI_FN(gboolean, gdk_event_get_scroll_direction, (const GdkEvent *event, GdkScrollDirection *direction)) /* (w) */
/* Меню окна (Window.SetMenu) */
NUI_FN(GtkWidget *, gtk_menu_bar_new, (void))                                                 /* (w) */
NUI_FN(GtkWidget *, gtk_event_box_new, (void))                                                /* (w) */
NUI_FN(void, gtk_event_box_set_visible_window, (GtkEventBox *event_box, gboolean visible_window)) /* (w) */
NUI_FN(GtkWidget *, gtk_menu_new, (void))
NUI_FN(GtkWidget *, gtk_menu_item_new_with_label, (const gchar *label))
NUI_FN(GtkWidget *, gtk_check_menu_item_new_with_label, (const gchar *label))
NUI_FN(GtkWidget *, gtk_separator_menu_item_new, (void))
NUI_FN(void, gtk_menu_item_set_submenu, (GtkMenuItem *menu_item, GtkWidget *submenu))
NUI_FN(void, gtk_menu_shell_append, (GtkMenuShell *menu_shell, GtkWidget *child))
NUI_FN(void, gtk_menu_item_set_label, (GtkMenuItem *menu_item, const gchar *label))
NUI_FN(const gchar *, gtk_menu_item_get_label, (GtkMenuItem *menu_item))                    /* тесты */
NUI_FN(void, gtk_menu_item_activate, (GtkMenuItem *menu_item))                               /* тесты */
NUI_FN(void, gtk_check_menu_item_set_active, (GtkCheckMenuItem *check_menu_item, gboolean is_active))
NUI_FN(gboolean, gtk_check_menu_item_get_active, (GtkCheckMenuItem *check_menu_item))
NUI_FN(void, gtk_accel_label_set_accel, (GtkAccelLabel *accel_label, guint accelerator_key,
       GdkModifierType accelerator_mods))
NUI_FN(void, gtk_accel_label_get_accel, (GtkAccelLabel *accel_label, guint *accelerator_key,
       GdkModifierType *accelerator_mods))                                                  /* тесты */
NUI_FN(void, gtk_widget_show_all, (GtkWidget *widget))                                       /* (w) */
NUI_FN(void, gtk_widget_get_size_request, (GtkWidget *widget, gint *width, gint *height))        /* (w) */
NUI_FN(void, gtk_box_pack_end, (GtkBox *box, GtkWidget *child, gboolean expand, gboolean fill, guint padding)) /* (w) */
/* OnDropFiles: окно принимает text/uri-list */
NUI_FN(void, gtk_drag_dest_set, (GtkWidget *widget, GtkDestDefaults flags, const GtkTargetEntry *targets,
       gint n_targets, GdkDragAction actions))                                               /* (w) */
NUI_FN(gchar **, gtk_selection_data_get_uris, (const GtkSelectionData *selection_data))      /* (w) */
NUI_FN(void, g_strfreev, (gchar **str_array))                                                 /* (w) */
NUI_FN(gchar *, g_filename_from_uri, (const gchar *uri, gchar **hostname, GError **error))    /* (w) */
/* TextView: OnEnter и подсказка в пустом поле */
NUI_FN(gboolean, gtk_text_view_im_context_filter_keypress, (GtkTextView *text_view, GdkEventKey *event)) /* (w) */
NUI_FN(gint, gtk_text_buffer_get_char_count, (GtkTextBuffer *buffer))                          /* (w) */
NUI_FN(void, gtk_text_buffer_get_start_iter, (GtkTextBuffer *buffer, GtkTextIter *iter))       /* (w) */
NUI_FN(void, gtk_text_view_get_iter_location, (GtkTextView *text_view, const GtkTextIter *iter, GdkRectangle *location)) /* (w) */
NUI_FN(void, gtk_text_view_buffer_to_window_coords, (GtkTextView *text_view, GtkTextWindowType win,
       gint buffer_x, gint buffer_y, gint *window_x, gint *window_y))                       /* (w) */
NUI_FN(GtkStateFlags, gtk_widget_get_state_flags, (GtkWidget *widget))                        /* (w) */
NUI_FN(void, gtk_style_context_get_color, (GtkStyleContext *context, GtkStateFlags state, GdkRGBA *color)) /* (w) */
NUI_FN(int, gtk_widget_get_allocated_width, (GtkWidget *widget))
NUI_FN(int, gtk_widget_get_allocated_height, (GtkWidget *widget))
NUI_FN(void, gtk_window_get_size, (GtkWindow *window, gint *width, gint *height))
NUI_FN(void, pango_layout_set_width, (PangoLayout *layout, int width))                       /* (w) */
NUI_FN(void, pango_layout_set_wrap, (PangoLayout *layout, PangoWrapMode wrap))               /* (w) */
NUI_FN(void, cairo_move_to, (cairo_t *cr, double x, double y))                                /* (w) */
NUI_FN(void, cairo_set_source_rgba, (cairo_t *cr, double red, double green, double blue, double alpha)) /* (w) */
NUI_FN(void, pango_cairo_show_layout, (cairo_t *cr, PangoLayout *layout))                    /* (w) */
NUI_FN(void, gtk_style_context_add_provider, (GtkStyleContext *context, GtkStyleProvider *provider, guint priority)) /* (w) */
NUI_FN(GtkCssProvider *, gtk_css_provider_new, (void))                                       /* (w) */
NUI_FN(gboolean, gtk_css_provider_load_from_data, (GtkCssProvider *css_provider, const gchar *data, gssize length,
       GError **error))                                                                     /* (w) */
NUI_FN(GtkWidget *, gtk_bin_get_child, (GtkBin *bin))

/* GtkWindow */
NUI_FN(GtkWidget *, gtk_window_new, (GtkWindowType type))
NUI_FN(void, gtk_window_set_title, (GtkWindow *window, const gchar *title))
NUI_FN(const gchar *, gtk_window_get_title, (GtkWindow *window))
NUI_FN(void, gtk_window_set_position, (GtkWindow *window, GtkWindowPosition position))
NUI_FN(void, gtk_window_present, (GtkWindow *window))
NUI_FN(void, gtk_window_set_default_size, (GtkWindow *window, gint width, gint height))
NUI_FN(void, gtk_window_resize, (GtkWindow *window, gint width, gint height))
NUI_FN(void, gtk_window_set_resizable, (GtkWindow *window, gboolean resizable))
NUI_FN(gboolean, gtk_window_get_resizable, (GtkWindow *window))
NUI_FN(void, gtk_window_get_default_size, (GtkWindow *window, gint *width, gint *height))
NUI_FN(void, gtk_window_set_default, (GtkWindow *window, GtkWidget *default_widget))
NUI_FN(void, gtk_window_set_icon_name, (GtkWindow *window, const gchar *name))
NUI_FN(void, gtk_window_maximize, (GtkWindow *window))
NUI_FN(void, gtk_window_unmaximize, (GtkWindow *window))
NUI_FN(void, gtk_window_fullscreen, (GtkWindow *window))
NUI_FN(void, gtk_window_unfullscreen, (GtkWindow *window))
NUI_FN(void, gtk_widget_set_can_default, (GtkWidget *widget, gboolean can_default))
NUI_FN(gboolean, gtk_widget_activate, (GtkWidget *widget)) /* тесты: SimEnter */
NUI_FN(void, gtk_window_close, (GtkWindow *window)) /* тесты: SimClose */

/* GtkLabel, GtkButton */
NUI_FN(GtkWidget *, gtk_label_new, (const gchar *str))
NUI_FN(void, gtk_label_set_text, (GtkLabel *label, const gchar *str))
NUI_FN(const gchar *, gtk_label_get_text, (GtkLabel *label))
NUI_FN(void, gtk_label_set_xalign, (GtkLabel *label, gfloat xalign))
NUI_FN(void, gtk_label_set_justify, (GtkLabel *label, GtkJustification jtype))
NUI_FN(void, gtk_label_set_ellipsize, (GtkLabel *label, PangoEllipsizeMode mode))
NUI_FN(void, gtk_label_set_line_wrap, (GtkLabel *label, gboolean wrap))
NUI_FN(void, gtk_label_set_selectable, (GtkLabel *label, gboolean setting))
NUI_FN(void, gtk_label_set_attributes, (GtkLabel *label, PangoAttrList *attrs))            /* (w) */
NUI_FN(PangoAttrList *, pango_attr_list_new, (void))                                        /* (w) */
NUI_FN(void, pango_attr_list_insert, (PangoAttrList *list, PangoAttribute *attr))           /* (w) */
NUI_FN(void, pango_attr_list_unref, (PangoAttrList *list))                                  /* (w) */
NUI_FN(PangoAttribute *, pango_attr_weight_new, (PangoWeight weight))                       /* (w) */
NUI_FN(PangoAttribute *, pango_attr_foreground_new, (guint16 red, guint16 green, guint16 blue)) /* (w) */
NUI_FN(PangoAttribute *, pango_attr_scale_new, (double scale_factor))                        /* (w) */
NUI_FN(GtkWidget *, gtk_button_new_with_label, (const gchar *label))
NUI_FN(void, gtk_button_set_label, (GtkButton *button, const gchar *label))
NUI_FN(const gchar *, gtk_button_get_label, (GtkButton *button))
NUI_FN(void, gtk_button_clicked, (GtkButton *button)) /* тесты: SimClick */
NUI_FN(GtkWidget *, gtk_check_button_new_with_label, (const gchar *label))
NUI_FN(void, gtk_toggle_button_set_active, (GtkToggleButton *toggle_button, gboolean is_active))
NUI_FN(gboolean, gtk_toggle_button_get_active, (GtkToggleButton *toggle_button))

/* GtkEntry */
NUI_FN(GtkWidget *, gtk_entry_new, (void))
NUI_FN(void, gtk_entry_set_text, (GtkEntry *entry, const gchar *text))
NUI_FN(const gchar *, gtk_entry_get_text, (GtkEntry *entry))
NUI_FN(void, gtk_entry_set_placeholder_text, (GtkEntry *entry, const gchar *text))
NUI_FN(void, gtk_entry_set_visibility, (GtkEntry *entry, gboolean visible))
NUI_FN(void, gtk_editable_set_editable, (GtkEditable *editable, gboolean is_editable))
NUI_FN(void, gtk_entry_set_width_chars, (GtkEntry *entry, gint n_chars))
NUI_FN(void, gtk_entry_set_activates_default, (GtkEntry *entry, gboolean setting))

/* GtkComboBoxText */
NUI_FN(GtkWidget *, gtk_combo_box_text_new, (void))
NUI_FN(GtkWidget *, gtk_combo_box_text_new_with_entry, (void))
NUI_FN(void, gtk_combo_box_text_append_text, (GtkComboBoxText *combo_box, const gchar *text))
NUI_FN(void, gtk_combo_box_text_remove_all, (GtkComboBoxText *combo_box))
NUI_FN(gint, gtk_combo_box_get_active, (GtkComboBox *combo_box))
NUI_FN(void, gtk_combo_box_set_active, (GtkComboBox *combo_box, gint index_))
NUI_FN(gchar *, gtk_combo_box_text_get_active_text, (GtkComboBoxText *combo_box)) /* g_free */

/* GtkScale, GtkRange, GtkProgressBar */
NUI_FN(GtkWidget *, gtk_scale_new_with_range, (GtkOrientation orientation, gdouble min, gdouble max, gdouble step))
NUI_FN(void, gtk_scale_set_digits, (GtkScale *scale, gint digits))
NUI_FN(void, gtk_range_set_value, (GtkRange *range, gdouble value))
NUI_FN(gdouble, gtk_range_get_value, (GtkRange *range))
NUI_FN(GtkWidget *, gtk_progress_bar_new, (void))
NUI_FN(void, gtk_progress_bar_set_fraction, (GtkProgressBar *pbar, gdouble fraction))
NUI_FN(gdouble, gtk_progress_bar_get_fraction, (GtkProgressBar *pbar))
NUI_FN(void, gtk_progress_bar_pulse, (GtkProgressBar *pbar))
NUI_FN(void, gtk_progress_bar_set_text, (GtkProgressBar *pbar, const gchar *text))
NUI_FN(void, gtk_progress_bar_set_show_text, (GtkProgressBar *pbar, gboolean show_text))

/* GtkTextView, GtkScrolledWindow */
NUI_FN(GtkWidget *, gtk_scrolled_window_new, (GtkAdjustment *hadjustment, GtkAdjustment *vadjustment))
NUI_FN(void, gtk_scrolled_window_set_policy, (GtkScrolledWindow *scrolled_window,
       GtkPolicyType hscrollbar_policy, GtkPolicyType vscrollbar_policy))
NUI_FN(void, gtk_scrolled_window_set_propagate_natural_width, (GtkScrolledWindow *scrolled_window, gboolean propagate))
NUI_FN(void, gtk_scrolled_window_set_propagate_natural_height, (GtkScrolledWindow *scrolled_window, gboolean propagate))
NUI_FN(GtkWidget *, gtk_text_view_new, (void))
NUI_FN(GtkTextBuffer *, gtk_text_view_get_buffer, (GtkTextView *text_view))
NUI_FN(void, gtk_text_buffer_set_text, (GtkTextBuffer *buffer, const gchar *text, gint len))
NUI_FN(void, gtk_text_view_set_editable, (GtkTextView *text_view, gboolean setting))
NUI_FN(void, gtk_text_view_set_cursor_visible, (GtkTextView *text_view, gboolean setting))
NUI_FN(void, gtk_text_view_set_monospace, (GtkTextView *text_view, gboolean monospace))
NUI_FN(void, gtk_text_view_set_wrap_mode, (GtkTextView *text_view, GtkWrapMode wrap_mode))
NUI_FN(void, gtk_text_view_set_left_margin, (GtkTextView *text_view, gint left_margin))
NUI_FN(void, gtk_text_view_set_right_margin, (GtkTextView *text_view, gint right_margin))
NUI_FN(void, gtk_text_view_set_top_margin, (GtkTextView *text_view, gint top_margin))
NUI_FN(void, gtk_text_view_set_bottom_margin, (GtkTextView *text_view, gint bottom_margin))
NUI_FN(void, gtk_text_buffer_get_start_iter, (GtkTextBuffer *buffer, GtkTextIter *iter))           /* (w) */
NUI_FN(void, gtk_text_buffer_get_end_iter, (GtkTextBuffer *buffer, GtkTextIter *iter))             /* (w) */
NUI_FN(void, gtk_text_buffer_get_iter_at_line, (GtkTextBuffer *buffer, GtkTextIter *iter, gint line_number)) /* (w) */
NUI_FN(gchar *, gtk_text_buffer_get_text, (GtkTextBuffer *buffer, const GtkTextIter *start,
       const GtkTextIter *end, gboolean include_hidden_chars))                                    /* (w) */
NUI_FN(void, gtk_text_buffer_insert, (GtkTextBuffer *buffer, GtkTextIter *iter, const gchar *text, gint len)) /* (w) */
NUI_FN(void, gtk_text_buffer_delete, (GtkTextBuffer *buffer, GtkTextIter *start, GtkTextIter *end)) /* (w) */
NUI_FN(gint, gtk_text_buffer_get_line_count, (GtkTextBuffer *buffer))                               /* (w) */
NUI_FN(gboolean, gtk_text_iter_starts_line, (const GtkTextIter *iter))                              /* (w) */
NUI_FN(GtkTextMark *, gtk_text_buffer_get_mark, (GtkTextBuffer *buffer, const gchar *name))         /* (w) */
NUI_FN(GtkTextMark *, gtk_text_buffer_create_mark, (GtkTextBuffer *buffer, const gchar *mark_name,
       const GtkTextIter *where, gboolean left_gravity))                                          /* (w) */
NUI_FN(void, gtk_text_buffer_move_mark, (GtkTextBuffer *buffer, GtkTextMark *mark, const GtkTextIter *where)) /* (w) */
NUI_FN(void, gtk_text_view_scroll_mark_onscreen, (GtkTextView *text_view, GtkTextMark *mark))       /* (w) */

/* Контейнеры */
NUI_FN(GtkWidget *, gtk_box_new, (GtkOrientation orientation, gint spacing))
NUI_FN(void, gtk_box_set_spacing, (GtkBox *box, gint spacing))
NUI_FN(void, gtk_box_pack_start, (GtkBox *box, GtkWidget *child, gboolean expand, gboolean fill, guint padding))
NUI_FN(GtkWidget *, gtk_grid_new, (void))
NUI_FN(void, gtk_grid_attach, (GtkGrid *grid, GtkWidget *child, gint left, gint top, gint width, gint height))
NUI_FN(void, gtk_grid_set_row_spacing, (GtkGrid *grid, guint spacing))
NUI_FN(void, gtk_grid_set_column_spacing, (GtkGrid *grid, guint spacing))
NUI_FN(GtkWidget *, gtk_separator_new, (GtkOrientation orientation))
NUI_FN(GtkWidget *, gtk_frame_new, (const gchar *label))
NUI_FN(void, gtk_frame_set_label, (GtkFrame *frame, const gchar *label))

/* Диалоги и оболочка */
NUI_FN(GtkWidget *, gtk_message_dialog_new, (GtkWindow *parent, GtkDialogFlags flags, GtkMessageType type,
       GtkButtonsType buttons, const gchar *message_format, ...))                                        /* (w) */
NUI_FN(void, gtk_message_dialog_format_secondary_text, (GtkMessageDialog *message_dialog,
       const gchar *message_format, ...))                                                                /* (w) */
NUI_FN(gint, gtk_dialog_run, (GtkDialog *dialog))                                                         /* (w) */
NUI_FN(void, gtk_dialog_response, (GtkDialog *dialog, gint response_id))                                  /* тесты */
NUI_FN_CONST(GType, gtk_dialog_get_type)                                                                  /* тесты */
NUI_FN_CONST(GType, gtk_file_chooser_get_type)                                                            /* тесты */
NUI_FN(GList *, gtk_window_list_toplevels, (void))                                                        /* тесты */
NUI_FN(GtkWidget *, gtk_widget_get_parent, (GtkWidget *widget))
NUI_FN(GtkWidget *, gtk_window_get_default_widget, (GtkWindow *window))                                   /* тесты */
NUI_FN(GList *, gtk_container_get_children, (GtkContainer *container))                                    /* (w) */
NUI_FN(GtkFileChooserNative *, gtk_file_chooser_native_new, (const gchar *title, GtkWindow *parent,
       GtkFileChooserAction action, const gchar *accept_label, const gchar *cancel_label))
NUI_FN(gint, gtk_native_dialog_run, (GtkNativeDialog *self))
NUI_FN(void, gtk_native_dialog_set_modal, (GtkNativeDialog *self, gboolean modal))
NUI_FN(gchar *, gtk_file_chooser_get_filename, (GtkFileChooser *chooser)) /* g_free */
NUI_FN(gboolean, gtk_file_chooser_set_filename, (GtkFileChooser *chooser, const char *filename))          /* тесты */
NUI_FN(gboolean, gtk_file_chooser_select_filename, (GtkFileChooser *chooser, const char *filename))       /* тесты */
NUI_FN_CONST(GType, gtk_tree_view_get_type)                                                               /* тесты */
NUI_FN_CONST(GType, gtk_text_view_get_type)                                                               /* тесты */
NUI_FN_CONST(GType, gtk_container_get_type)                                                               /* тесты */
NUI_FN(void, gtk_container_forall, (GtkContainer *container, GtkCallback callback, gpointer callback_data)) /* тесты */
NUI_FN(gboolean, gtk_widget_get_mapped, (GtkWidget *widget))
NUI_FN(GtkTreeSelection *, gtk_tree_view_get_selection, (GtkTreeView *tree_view))                         /* тесты */
NUI_FN(GList *, gtk_tree_selection_get_selected_rows, (GtkTreeSelection *selection, GtkTreeModel **model)) /* тесты */
NUI_FN(void, gtk_tree_selection_select_path, (GtkTreeSelection *selection, GtkTreePath *path))            /* тесты */
NUI_FN(void, gtk_tree_path_free, (GtkTreePath *path))                                                     /* тесты */
NUI_FN(void, gtk_file_chooser_set_select_multiple, (GtkFileChooser *chooser, gboolean select_multiple))
NUI_FN(GSList *, gtk_file_chooser_get_filenames, (GtkFileChooser *chooser)) /* g_free каждую, g_slist_free */
NUI_FN(void, gtk_file_chooser_set_current_name, (GtkFileChooser *chooser, const gchar *name))
NUI_FN(void, gtk_file_chooser_set_do_overwrite_confirmation, (GtkFileChooser *chooser, gboolean do_overwrite_confirmation))
NUI_FN(GtkFileFilter *, gtk_file_filter_new, (void))
NUI_FN(void, gtk_file_filter_set_name, (GtkFileFilter *filter, const gchar *name))
NUI_FN(void, gtk_file_filter_add_pattern, (GtkFileFilter *filter, const gchar *pattern))
NUI_FN(void, gtk_file_chooser_add_filter, (GtkFileChooser *chooser, GtkFileFilter *filter))
NUI_FN(gboolean, gtk_show_uri_on_window, (GtkWindow *parent, const char *uri, guint32 timestamp, GError **error))
NUI_FN(GDBusConnection *, g_bus_get_sync, (GBusType bus_type, GCancellable *cancellable, GError **error))  /* (w) */
NUI_FN(GVariant *, g_variant_new, (const gchar *format_string, ...))                                       /* (w) */
NUI_FN(void, g_dbus_connection_call, (GDBusConnection *connection, const gchar *bus_name,
       const gchar *object_path, const gchar *interface_name, const gchar *method_name,
       GVariant *parameters, const GVariantType *reply_type, GDBusCallFlags flags, gint timeout_msec,
       GCancellable *cancellable, GAsyncReadyCallback callback, gpointer user_data))                      /* (w) */
NUI_FN(GVariant *, g_dbus_connection_call_finish, (GDBusConnection *connection, GAsyncResult *res, GError **error)) /* (w) */
NUI_FN(void, g_variant_unref, (GVariant *value))                                                           /* (w) */

/* ---------- 5. Наши обёртки (wrap_linux.c) и трамплины (export_linux.go) ----------
   Объявляются в ОБОИХ режимах: в режиме проверки wrap_linux.c компилируется против
   настоящих заголовков (tools/check-protos.sh, шаг 2), и без этих объявлений трамплины
   остались бы необъявленными. */
gulong nui_connect(gpointer obj, const gchar *sig, uintptr_t ud);
gulong nui_connect_ret(gpointer obj, const gchar *sig, uintptr_t ud);
guint  nui_timeout(guint ms, uintptr_t ud);
guint  nui_idle(uintptr_t ud);
guint  nui_watch(int fd, unsigned cond, uintptr_t ud);
void   nui_label_style(GtkLabel *l, int set_bold, int bold, int set_color, unsigned color, int scale_pct);
void   nui_css(GtkWidget *w, int what, unsigned value);
void   nui_css_background(GtkWidget *w, unsigned color, int panel);
void   nui_set_padding(GtkWidget *w, unsigned px);
GtkClipboard *nui_clipboard(void);
void   nui_container_clear(GtkContainer *c);
void   nui_paint_begin(cairo_t *cr, GtkWidget *w);
void   nui_paint_theme_color(cairo_t *cr, GtkWidget *w);
unsigned nui_pixel_at(GtkWidget *w, int x, int y); /* только с NUI_GUI_TEST */
void   nui_paint_color(cairo_t *cr, unsigned rgb);
void   nui_paint_line(cairo_t *cr, double x1, double y1, double x2, double y2);
void   nui_paint_rect(cairo_t *cr, double x, double y, double w, double h, int fill);
void   nui_paint_ellipse(cairo_t *cr, double x, double y, double w, double h, int fill);
void   nui_paint_text(cairo_t *cr, GtkWidget *w, double x, double y, const gchar *s);
void   nui_paint_text_size(GtkWidget *w, const gchar *s, int out[2]);
int    nui_mouse_decode(GdkEvent *e, double xy[2], int info[3]); /* вид; info: кнопка, Mod*, шаг колеса */
void   nui_sim_mouse(GtkWidget *w, int type, double x, double y, unsigned button, unsigned state); /* только с NUI_GUI_TEST */
GtkWidget *nui_window_menubar(GtkWindow *w);
GtkContainer *nui_window_content(GtkWindow *w);
void   nui_ctx_set(GtkWidget *w, GtkWidget *menu);
GtkWidget *nui_ctx_owner_at(GtkWidget *top, int x, int y); /* только с NUI_GUI_TEST: чьё меню у точки окна */
void   nui_window_menubar_resize(GtkWindow *w);
GtkWidget *nui_tabs_new(uintptr_t ud);
void   nui_radio_add(GtkBox *box, const gchar *text, uintptr_t ud);
GtkWidget *nui_radio_nth(GtkBox *box, int i); /* кнопка i; 0 — нет такой */
int    nui_radio_selected(GtkBox *box);       /* номер включённой; -1 — кнопок нет */
GtkWidget *nui_image_new(void);
GtkWidget *nui_list_new(GtkWidget **view, uintptr_t activate_ud);
void   nui_list_column(GtkTreeView *view, const gchar *title, int index);
void   nui_list_set_rows(GtkTreeView *view, int rows, uintptr_t key);
int    nui_list_selected(GtkTreeView *view);
void   nui_list_select(GtkTreeView *view, int row);
guint  nui_key_by_code(GdkEventKey *e);
void   nui_text_enter(GtkTextView *view, int on, uintptr_t ud);
void   nui_text_placeholder(GtkTextView *view, const gchar *s);
void   nui_drop_enable(GtkWidget *w, uintptr_t ud);
void   nui_drop_uris(GtkWidget *w, gchar **uris, uintptr_t ud);
int    nui_sim_key(GtkWidget *w, guint keyval, guint code_of, guint state); /* только с NUI_GUI_TEST */
char  *nui_list_cell_text(GtkTreeView *view, int row, int col); /* только с NUI_GUI_TEST; g_free */
void   nui_sim_list_activate(GtkTreeView *view, int row);       /* только с NUI_GUI_TEST */
int    nui_image_load(GtkWidget *w, const guchar *data, gsize len, int wh[2]);
void   nui_image_show(GtkWidget *w, int width, int height);
char  *nui_text_get(GtkTextBuffer *b);
void   nui_text_append(GtkTextView *v, const gchar *s, int len, int autoscroll, int max_lines);
int    nui_message(GtkWindow *parent, int type, int buttons, const gchar *title, const gchar *text);
int    nui_show_items(const gchar *uri, const gchar *dir_uri);
void   nui_allocation(GtkWidget *w, int out[4]); /* только с NUI_GUI_TEST */
void   nui_preferred(GtkWidget *w, int out[4]);  /* только с NUI_GUI_TEST */
void   nui_window_pos(GtkWidget *w, int out[4]); /* только с NUI_GUI_TEST */
int    nui_sim_dialog_response(int chooser, int response); /* только с NUI_GUI_TEST */
int    nui_sim_file_chooser_set(const gchar *path);         /* только с NUI_GUI_TEST */
int    nui_sim_file_chooser_add(const gchar *path);         /* только с NUI_GUI_TEST */
int    nui_dialog_open(int chooser);                        /* только с NUI_GUI_TEST */

void nuiSignal(void *obj, uintptr_t ud);
gchar *nuiListCell(uintptr_t key, int row, int col); /* текст ячейки List, g_malloc */
void nuiListDrop(uintptr_t key);                     /* модель List уничтожена */
int  nuiSignalRet(void *obj, void *arg, uintptr_t ud);
int  nuiSource(uintptr_t ud);
int  nuiWatch(int fd, unsigned cond, uintptr_t ud);

#endif
