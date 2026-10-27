/* wrap_linux.c — то, что Go не может вызвать сам: макрос g_signal_connect,
   функции с переменным числом аргументов, структуры на стеке, указатели на колбэки.
   Трамплины приводятся к типам колбэков GLib: по ABI uintptr_t и gpointer
   передаются одинаково (на том же держится макрос G_CALLBACK). */
#include "gtk3_linux.h"

gulong nui_connect(gpointer obj, const gchar *sig, uintptr_t ud) {
	return g_signal_connect_data(obj, sig, (GCallback)nuiSignal, (gpointer)ud, 0, 0);
}

gulong nui_connect_ret(gpointer obj, const gchar *sig, uintptr_t ud) {
	return g_signal_connect_data(obj, sig, (GCallback)nuiSignalRet, (gpointer)ud, 0, 0);
}

guint nui_timeout(guint ms, uintptr_t ud) {
	return g_timeout_add_full(NUI_G_PRIORITY_DEFAULT, ms, (GSourceFunc)nuiSource, (gpointer)ud, 0);
}

/* Потокобезопасна: GLib разрешает g_idle_add_full из любого потока. */
guint nui_idle(uintptr_t ud) {
	return g_idle_add_full(NUI_G_PRIORITY_DEFAULT, (GSourceFunc)nuiSource, (gpointer)ud, 0);
}

/* HUP, ERR и NVAL подписываются всегда, как у poll (api.md, §3). */
guint nui_watch(int fd, unsigned cond, uintptr_t ud) {
	cond |= NUI_G_IO_HUP | NUI_G_IO_ERR | NUI_G_IO_NVAL;
	return g_unix_fd_add_full(NUI_G_PRIORITY_DEFAULT, fd, (GIOCondition)cond,
	                          (GUnixFDSourceFunc)nuiWatch, (gpointer)ud, 0);
}

/* Жирный шрифт и цвет метки — атрибуты Pango. Состояние хранится в самой метке
   (g_object_set_data, число, а не указатель Go), поэтому у nui.Label нет лишних полей
   и приложения без оформления не растут ни на байт. Упаковка:
   биты 0–23 — цвет 0xRRGGBB, бит 24 — без цвета (цвет темы), бит 25 — жирный,
   бит 26 — значение задано. color == 1<<24 (nui.NoColor) — цвет темы. */
#define NUI_STYLE_NOCOLOR (1u << 24)
#define NUI_STYLE_BOLD    (1u << 25)
#define NUI_STYLE_SET     (1u << 26)
void nui_label_style(GtkLabel *l, int set_bold, int bold, int set_color, unsigned color, int scale_pct) {
	unsigned st = (unsigned)(uintptr_t)g_object_get_data((GObject *)l, "nui-style");
	if (!st)
		st = NUI_STYLE_SET | NUI_STYLE_NOCOLOR;
	if (set_bold)
		st = bold ? (st | NUI_STYLE_BOLD) : (st & ~NUI_STYLE_BOLD);
	if (set_color)
		st = (st & (NUI_STYLE_SET | NUI_STYLE_BOLD)) | (color & (NUI_STYLE_NOCOLOR | 0xffffffu));
	g_object_set_data((GObject *)l, "nui-style", (gpointer)(uintptr_t)st);

	PangoAttrList *a = pango_attr_list_new();
	pango_attr_list_insert(a, pango_attr_weight_new((st & NUI_STYLE_BOLD) ? NUI_PANGO_WEIGHT_BOLD : NUI_PANGO_WEIGHT_NORMAL));
	if (!(st & NUI_STYLE_NOCOLOR)) {
		unsigned r = (st >> 16) & 255, g = (st >> 8) & 255, b = st & 255;
		pango_attr_list_insert(a, pango_attr_foreground_new(r * 257, g * 257, b * 257));
	}
	/* Масштаб шрифта (SetFontScale) — в процентах, хранится отдельно; 0 — не менять. */
	if (scale_pct > 0)
		g_object_set_data((GObject *)l, "nui-scale", (gpointer)(uintptr_t)scale_pct);
	int pct = (int)(uintptr_t)g_object_get_data((GObject *)l, "nui-scale");
	if (pct > 0 && pct != 100)
		pango_attr_list_insert(a, pango_attr_scale_new(pct / 100.0));
	gtk_label_set_attributes(l, a);
	pango_attr_list_unref(a);
}

/* --- Цвета и фон (ADR-19): свой GtkCssProvider на виджет. Состояние — в данных объекта
   (значение + 1, 0 — не задано): у Go-структур nui нет лишних полей. ---------------------- */
static const char *const css_keys[3] = {"nui-fg", "nui-bg", "nui-pad"};

static unsigned css_get(GtkWidget *w, int what) {
	return (unsigned)(uintptr_t)g_object_get_data((GObject *)w, css_keys[what]);
}

static char *css_put(char *p, const char *s) {
	while (*s)
		*p++ = *s++;
	return p;
}

static char *css_hex(char *p, unsigned v) {
	static const char d[] = "0123456789abcdef";
	*p++ = '#';
	for (int i = 20; i >= 0; i -= 4)
		*p++ = d[(v >> i) & 15];
	return p;
}

static char *css_dec(char *p, unsigned v) {
	char t[12];
	int n = 0;
	do
		t[n++] = (char)('0' + v % 10);
	while ((v /= 10) && n < 11);
	while (n)
		*p++ = t[--n];
	return p;
}

/* Цвета одного правила: текст и фон (фон — и вместо градиента темы). */
static char *css_colors(char *p, unsigned fg, unsigned bg) {
	if (fg && !((fg - 1) & NUI_STYLE_NOCOLOR)) {
		p = css_put(p, "color:");
		p = css_hex(p, fg - 1);
		p = css_put(p, ";");
	}
	if (bg && !((bg - 1) & NUI_STYLE_NOCOLOR)) {
		p = css_put(p, "background-color:");
		p = css_hex(p, bg - 1);
		p = css_put(p, ";background-image:none;");
	}
	return p;
}

/* nui_css: what — 0 цвет текста, 1 фон, 2 отступ панели (Padding цветного Box/Grid);
   цвет — 0xRRGGBB или NUI_STYLE_NOCOLOR (цвет темы). Подузел text — поле TextView. */
void nui_css(GtkWidget *w, int what, unsigned value) {
	g_object_set_data((GObject *)w, css_keys[what], (gpointer)(uintptr_t)(value + 1));
	unsigned fg = css_get(w, 0), bg = css_get(w, 1), pad = css_get(w, 2);
	char css[192], *p = css;
	p = css_put(p, "*{");
	p = css_colors(p, fg, bg);
	if (pad) {
		p = css_put(p, "padding:");
		p = css_dec(p, pad - 1);
		p = css_put(p, "px;");
	}
	p = css_put(p, "}* text{");
	p = css_colors(p, fg, bg);
	p = css_put(p, "}");
	*p = 0;
	GtkCssProvider *prov = g_object_get_data((GObject *)w, "nui-css");
	if (!prov) {
		prov = gtk_css_provider_new();
		gtk_style_context_add_provider(gtk_widget_get_style_context(w), (GtkStyleProvider *)prov,
			NUI_GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
		g_object_set_data_full((GObject *)w, "nui-css", prov, g_object_unref);
	}
	gtk_css_provider_load_from_data(prov, css, -1, 0);
}

/* Фон. panel — Box или Grid: их border-width лежит вне фона, поэтому у цветной панели он
   становится CSS padding (раскладка детей та же, отступ закрашен); без цвета — обратно. */
void nui_css_background(GtkWidget *w, unsigned color, int panel) {
	if (panel) {
		unsigned pad = css_get(w, 2);
		if (!(color & NUI_STYLE_NOCOLOR) && !pad) {
			nui_css(w, 2, gtk_container_get_border_width((GtkContainer *)w));
			gtk_container_set_border_width((GtkContainer *)w, 0);
		} else if ((color & NUI_STYLE_NOCOLOR) && pad) {
			g_object_set_data((GObject *)w, css_keys[2], 0);
			gtk_container_set_border_width((GtkContainer *)w, pad - 1);
		}
	}
	nui_css(w, 1, color);
}

/* Padding: у цветной панели — CSS padding (закрашивается вместе с фоном), иначе border-width. */
void nui_set_padding(GtkWidget *w, unsigned px) {
	if (css_get(w, 2))
		nui_css(w, 2, px);
	else
		gtk_container_set_border_width((GtkContainer *)w, px);
}

/* Буфер обмена (CLIPBOARD, не PRIMARY). GDK_SELECTION_CLIPBOARD = _GDK_MAKE_ATOM (69),
   gdkselection.h; в режиме проверки берётся настоящий макрос. */
GtkClipboard *nui_clipboard(void) {
#ifdef NUI_CHECK
	return gtk_clipboard_get(GDK_SELECTION_CLIPBOARD);
#else
	return gtk_clipboard_get((GdkAtom)(uintptr_t)69);
#endif
}

/* Box.Clear, Grid.Clear: уничтожить всех детей (destroy → EvDestroy каждому). Список —
   копия, поэтому уничтожать по ходу обхода безопасно. */
void nui_container_clear(GtkContainer *c) {
	GList *kids = gtk_container_get_children(c);
	for (GList *it = kids; it; it = it->next)
		gtk_widget_destroy(it->data);
	g_list_free(kids);
}

/* ---------- Меню окна ----------
   Окно без меню — как было: GtkWindow (отступ — его border-width) → содержимое. Первое
   SetMenu перестраивает его: GtkWindow (без отступа) → ряд [строка меню, inner], inner —
   GtkEventBox без своего окна (прозрачный: фон окна виден) с отступом окна, в нём
   содержимое. Так меню во всю ширину, а отступ и раскладка содержимого прежние. inner,
   как и окно, — GtkBin: WindowSetChild одинаков для обоих. */

/* Куда класть содержимое окна и чей border-width — его отступ. */
GtkContainer *nui_window_content(GtkWindow *w) {
	GtkWidget *inner = g_object_get_data((GObject *)w, "nui-inner");
	return inner ? (GtkContainer *)inner : (GtkContainer *)w;
}

/* Новая пустая строка меню окна; прежняя уничтожается вместе с пунктами. */
GtkWidget *nui_window_menubar(GtkWindow *w) {
	GtkWidget *outer = g_object_get_data((GObject *)w, "nui-outer");
	if (!outer) {
		GtkWidget *child = gtk_bin_get_child((GtkBin *)w);
		GtkWidget *inner = gtk_event_box_new();
		gtk_event_box_set_visible_window((GtkEventBox *)inner, 0);
		outer = gtk_box_new(NUI_GTK_ORIENTATION_VERTICAL, 0);
		gtk_container_set_border_width((GtkContainer *)inner, gtk_container_get_border_width((GtkContainer *)w));
		gtk_container_set_border_width((GtkContainer *)w, 0);
		if (child) {
			g_object_ref(child);
			gtk_container_remove((GtkContainer *)w, child);
			gtk_container_add((GtkContainer *)inner, child);
			g_object_unref(child);
		}
		gtk_box_pack_end((GtkBox *)outer, inner, 1, 1, 0);
		gtk_container_add((GtkContainer *)w, outer);
		gtk_widget_show(inner);
		gtk_widget_show(outer);
		g_object_set_data_full((GObject *)w, "nui-outer", outer, 0);
		g_object_set_data_full((GObject *)w, "nui-inner", inner, 0);
	}
	GtkWidget *old = g_object_get_data((GObject *)w, "nui-menubar");
	if (old)
		gtk_widget_destroy(old);
	GtkWidget *bar = gtk_menu_bar_new();
	gtk_box_pack_start((GtkBox *)outer, bar, 0, 0, 0);
	g_object_set_data_full((GObject *)w, "nui-menubar", bar, 0);
	return bar;
}

/* Строка меню появилась, сменилась или убрана: окно растёт или уменьшается на разницу
   её высоты, и область под меню остаётся прежней — как клиентская область на Windows.
   Касается размера по умолчанию (окно ещё не показано), текущего размера и size request
   нерастягиваемого окна. Высота строки — в данных окна "nui-menubar-h": её читают
   WindowSize и WindowSetSize (программам без меню измерять нечего). */
void nui_window_menubar_resize(GtkWindow *w) {
	int old = (int)(uintptr_t)g_object_get_data((GObject *)w, "nui-menubar-h"), h = 0;
	GtkWidget *bar = g_object_get_data((GObject *)w, "nui-menubar");
	if (bar) {
		int min = 0;
		gtk_widget_show_all(bar);
		gtk_widget_get_preferred_height(bar, &min, &h);
	}
	g_object_set_data_full((GObject *)w, "nui-menubar-h", (gpointer)(uintptr_t)h, 0);
	int d = h - old, ww, wh;
	if (!d)
		return;
	gtk_window_get_default_size(w, &ww, &wh);
	if (wh > 0)
		gtk_window_set_default_size(w, ww, wh + d);
	gtk_widget_get_size_request((GtkWidget *)w, &ww, &wh);
	if (wh > 0)
		gtk_widget_set_size_request((GtkWidget *)w, ww, wh + d);
	if (gtk_widget_get_visible((GtkWidget *)w)) {
		gtk_window_get_size(w, &ww, &wh);
		gtk_window_resize(w, ww, wh + d);
	}
}

/* ---------- Canvas: рисование cairo в логических пикселях ----------
   Масштаб экрана cairo учитывает сам (поверхность окна с device scale). Каждая фигура —
   с нового пути: иначе линия соединилась бы с предыдущей. */

/* Цвет текста темы. */
void nui_paint_theme_color(cairo_t *cr, GtkWidget *w) {
	GdkRGBA c;
	gtk_style_context_get_color(gtk_widget_get_style_context(w), gtk_widget_get_state_flags(w), &c);
	cairo_set_source_rgba(cr, c.red, c.green, c.blue, c.alpha);
}

/* Начало draw: цвет — цвет текста темы, толщина линии — 1. */
void nui_paint_begin(cairo_t *cr, GtkWidget *w) {
	nui_paint_theme_color(cr, w);
	cairo_set_line_width(cr, 1);
}

void nui_paint_color(cairo_t *cr, unsigned rgb) {
	cairo_set_source_rgb(cr, (rgb >> 16 & 0xFF) / 255.0, (rgb >> 8 & 0xFF) / 255.0, (rgb & 0xFF) / 255.0);
}

void nui_paint_line(cairo_t *cr, double x1, double y1, double x2, double y2) {
	cairo_new_path(cr);
	cairo_move_to(cr, x1, y1);
	cairo_line_to(cr, x2, y2);
	cairo_stroke(cr);
}

void nui_paint_rect(cairo_t *cr, double x, double y, double w, double h, int fill) {
	cairo_new_path(cr);
	cairo_rectangle(cr, x, y, w, h);
	if (fill)
		cairo_fill(cr);
	else
		cairo_stroke(cr);
}

/* Эллипс, вписанный в прямоугольник: окружность в масштабе (толщина линии — после
   restore, поэтому одинаковая по всей линии). */
void nui_paint_ellipse(cairo_t *cr, double x, double y, double w, double h, int fill) {
	if (w <= 0 || h <= 0)
		return;
	cairo_new_path(cr);
	cairo_save(cr);
	cairo_translate(cr, x + w / 2, y + h / 2);
	cairo_scale(cr, w / 2, h / 2);
	cairo_arc(cr, 0, 0, 1, 0, 6.283185307179586);
	cairo_restore(cr);
	if (fill)
		cairo_fill(cr);
	else
		cairo_stroke(cr);
}

/* Текст шрифтом интерфейса, (x, y) — левый верхний угол. */
void nui_paint_text(cairo_t *cr, GtkWidget *w, double x, double y, const gchar *s) {
	PangoLayout *l = gtk_widget_create_pango_layout(w, s);
	cairo_new_path(cr);
	cairo_move_to(cr, x, y);
	pango_cairo_show_layout(cr, l);
	g_object_unref(l);
}

void nui_paint_text_size(GtkWidget *w, const gchar *s, int out[2]) {
	PangoLayout *l = gtk_widget_create_pango_layout(w, s);
	pango_layout_get_pixel_size(l, &out[0], &out[1]);
	g_object_unref(l);
}

/* Событие мыши → вид (1 нажатие, 2 отпускание, 3 движение, 4 колесо; 0 — пропустить:
   двойное и тройное нажатие GTK присылает лишним событием после второго обычного),
   координаты и info: кнопка, модификаторы nui, колесо (−1 вверх, +1 вниз, −2 влево,
   +2 вправо). */
int nui_mouse_decode(GdkEvent *e, double xy[2], int info[3]) {
	int kind = 0;
	guint button = 0;
	GdkModifierType st = 0;
	GdkScrollDirection dir = 0;
	switch (gdk_event_get_event_type(e)) {
	case NUI_GDK_BUTTON_PRESS:
		kind = 1;
		break;
	case NUI_GDK_BUTTON_RELEASE:
		kind = 2;
		break;
	case NUI_GDK_MOTION_NOTIFY:
		kind = 3;
		break;
	case NUI_GDK_SCROLL:
		kind = 4;
		break;
	default:
		return 0;
	}
	gdk_event_get_coords(e, &xy[0], &xy[1]);
	gdk_event_get_button(e, &button);
	gdk_event_get_state(e, &st);
	info[0] = (int)button;
	info[1] = (st & NUI_GDK_SHIFT_MASK ? 1 : 0) | (st & NUI_GDK_CONTROL_MASK ? 2 : 0) | (st & NUI_GDK_MOD1_MASK ? 4 : 0);
	info[2] = 0;
	if (kind == 4) {
		if (!gdk_event_get_scroll_direction(e, &dir))
			return 0; /* плавная прокрутка не запрошена — сюда не приходит */
		int d = (int)dir;
		info[2] = d == NUI_GDK_SCROLL_UP ? -1 : d == NUI_GDK_SCROLL_DOWN ? 1 : d == NUI_GDK_SCROLL_LEFT ? -2 : 2;
	}
	return kind;
}

/* ---------- Контекстное меню ----------
   Меню виджета — GtkMenu в его данных "nui-ctx", прикреплено к нему (уничтожится с ним).
   Виджеты со своим окном событий (поля, список, холст) ловят правый щелчок сами — раньше
   своего обычного меню. У подписей и рядов в GTK3 окна нет: щелчок доходит до окна
   верхнего уровня, и оно ищет самый глубокий виджет под указателем. Меню берётся у него
   или у ближайшего контейнера выше. */

static GtkWidget *ctx_find(GtkWidget *w) {
	for (; w; w = gtk_widget_get_parent(w))
		if (g_object_get_data((GObject *)w, "nui-ctx"))
			return w;
	return 0;
}

static gboolean ctx_popup(GtkWidget *owner, GdkEvent *e) {
	GtkMenu *m = owner ? g_object_get_data((GObject *)owner, "nui-ctx") : 0;
	if (!m)
		return 0;
	if (e)
		gtk_menu_popup_at_pointer(m, e);
	else
		gtk_menu_popup_at_widget(m, owner, 7 /* SOUTH_WEST */, 1 /* NORTH_WEST */, 0);
	return 1;
}

typedef struct {
	GtkWidget *top;
	int x, y;
	GtkWidget *hit;
} CtxPick;

static void ctx_pick(GtkWidget *w, gpointer data) {
	CtxPick *p = data;
	GtkAllocation a;
	int x, y;
	if (!gtk_widget_get_mapped(w) || !gtk_widget_translate_coordinates(p->top, w, p->x, p->y, &x, &y))
		return;
	gtk_widget_get_allocation(w, &a);
	if (x < 0 || y < 0 || x >= a.width || y >= a.height)
		return;
	p->hit = w; /* глубже — точнее */
	if (g_type_check_instance_is_a((GTypeInstance *)w, gtk_container_get_type()))
		gtk_container_forall((GtkContainer *)w, ctx_pick, data);
}

/* Владелец меню у точки (x, y) окна top: самый глубокий виджет под ней или контейнер выше. */
GtkWidget *nui_ctx_owner_at(GtkWidget *top, int x, int y) {
	CtxPick p = {top, x, y, 0};
	gtk_container_forall((GtkContainer *)top, ctx_pick, &p);
	return ctx_find(p.hit ? p.hit : top);
}

static gboolean ctx_window_press(GtkWidget *top, GdkEventButton *e, gpointer d) {
	(void)d;
	if ((int)e->type != NUI_GDK_BUTTON_PRESS || e->button != 3)
		return 0;
	int ox = 0, oy = 0;
	gdk_window_get_origin(gtk_widget_get_window(top), &ox, &oy);
	return ctx_popup(nui_ctx_owner_at(top, (int)e->x_root - ox, (int)e->y_root - oy), (GdkEvent *)e);
}

/* Правый щелчок по виджету со своим окном событий. В списке сначала выбирается строка
   под указателем — как в файловых менеджерах. */
static gboolean ctx_press(GtkWidget *w, GdkEventButton *e, gpointer d) {
	(void)d;
	if ((int)e->type != NUI_GDK_BUTTON_PRESS || e->button != 3)
		return 0;
	if (g_type_check_instance_is_a((GTypeInstance *)w, gtk_tree_view_get_type())) {
		GtkTreePath *p = 0;
		if (gtk_tree_view_get_path_at_pos((GtkTreeView *)w, (int)e->x, (int)e->y, &p, 0, 0, 0) && p) {
			gtk_tree_view_set_cursor((GtkTreeView *)w, p, 0, 0);
			gtk_tree_path_free(p);
		}
	}
	return ctx_popup(ctx_find(w), (GdkEvent *)e);
}

/* Клавиша «Меню» или Shift+F10 у виджета в фокусе. */
static gboolean ctx_key(GtkWidget *w, gpointer d) {
	(void)d;
	return ctx_popup(ctx_find(w), 0);
}

static void ctx_anchor(GtkWidget *w, GtkWidget *prev, gpointer d) {
	(void)prev;
	(void)d;
	GtkWidget *top = gtk_widget_get_toplevel(w);
	if (!gtk_widget_is_toplevel(top) || g_object_get_data((GObject *)top, "nui-ctx-top"))
		return;
	g_object_set_data_full((GObject *)top, "nui-ctx-top", (gpointer)1, 0);
	g_signal_connect_data(top, "button-press-event", (GCallback)ctx_window_press, 0, 0, 0);
}

static void ctx_hook(GtkWidget *w) {
	g_signal_connect_data(w, "button-press-event", (GCallback)ctx_press, 0, 0, 0);
	g_signal_connect_data(w, "popup-menu", (GCallback)ctx_key, 0, 0, 0);
}

/* menu == 0 — убрать меню. */
void nui_ctx_set(GtkWidget *w, GtkWidget *menu) {
	GtkWidget *old = g_object_get_data((GObject *)w, "nui-ctx");
	g_object_set_data_full((GObject *)w, "nui-ctx", menu, 0);
	if (old)
		gtk_widget_destroy(old);
	if (!menu)
		return;
	gtk_menu_attach_to_widget((GtkMenu *)menu, w, 0);
	if (g_object_get_data((GObject *)w, "nui-ctx-hooked"))
		return;
	g_object_set_data_full((GObject *)w, "nui-ctx-hooked", (gpointer)1, 0);
	if (gtk_widget_is_toplevel(w)) {
		ctx_anchor(w, 0, 0);
		return;
	}
	ctx_hook(w);
	if (g_type_check_instance_is_a((GTypeInstance *)w, gtk_scrolled_window_get_type())) {
		GtkWidget *inner = gtk_bin_get_child((GtkBin *)w); /* TextView, List: события — у них */
		if (inner)
			ctx_hook(inner);
	}
	g_signal_connect_data(w, "hierarchy-changed", (GCallback)ctx_anchor, 0, 0, 0);
	ctx_anchor(w, 0, 0);
}

/* ---------- Вкладки (Tabs): GtkNotebook ----------
   switch-page приходит и на программные переключения. Пользователь переключает только
   с видимой страницы; GTK сам уходит со страницы, которую прячут (gtk_widget_hide), —
   это не действие пользователя (ADR-8): до переключения запоминаем, видна ли текущая,
   после — EvChange, только если была видна. Add, Remove, SetCurrent глушит nui. */

static void tabs_before(GtkNotebook *nb, GtkWidget *page, guint num, gpointer ud) {
	(void)page;
	(void)num;
	(void)ud;
	gint cur = gtk_notebook_get_current_page(nb);
	GtkWidget *old = cur < 0 ? 0 : gtk_notebook_get_nth_page(nb, cur);
	g_object_set_data((GObject *)nb, "nui-user", (gpointer)(uintptr_t)(old && gtk_widget_get_visible(old)));
}

static void tabs_after(GtkNotebook *nb, GtkWidget *page, guint num, gpointer ud) {
	(void)page;
	(void)num;
	if (g_object_get_data((GObject *)nb, "nui-user"))
		nuiSignal(nb, (uintptr_t)ud);
}

/* Ярлыки, которые не помещаются, прокручиваются стрелками, как на Windows. */
GtkWidget *nui_tabs_new(uintptr_t ud) {
	GtkWidget *nb = gtk_notebook_new();
	gtk_notebook_set_scrollable((GtkNotebook *)nb, 1);
	g_signal_connect_data(nb, "switch-page", (GCallback)tabs_before, 0, 0, 0);
	g_signal_connect_data(nb, "switch-page", (GCallback)tabs_after, (gpointer)ud, 0, 1 /* G_CONNECT_AFTER */);
	return nb;
}

/* ---------- Переключатель (Radio): GtkRadioButton одной группы в GtkBox ----------
   toggled приходит и кнопке, которая выключилась, и той, что включилась: EvChange — только
   от включившейся. */

static void radio_toggled(GtkToggleButton *b, gpointer ud) {
	if (gtk_toggle_button_get_active(b))
		nuiSignal(b, (uintptr_t)ud);
}

/* Первая кнопка ряда — член группы для следующих; первая сразу включена. */
void nui_radio_add(GtkBox *box, const gchar *text, uintptr_t ud) {
	GList *kids = gtk_container_get_children((GtkContainer *)box);
	GtkWidget *b = gtk_radio_button_new_with_label_from_widget(kids ? (GtkRadioButton *)kids->data : 0, text);
	g_list_free(kids);
	gtk_widget_show(b);
	gtk_box_pack_start(box, b, 0, 0, 0);
	g_signal_connect_data(b, "toggled", (GCallback)radio_toggled, (gpointer)ud, 0, 0);
}

GtkWidget *nui_radio_nth(GtkBox *box, int i) {
	GList *kids = gtk_container_get_children((GtkContainer *)box);
	GList *l = i >= 0 ? g_list_nth(kids, (guint)i) : 0;
	GtkWidget *b = l ? l->data : 0;
	g_list_free(kids);
	return b;
}

int nui_radio_selected(GtkBox *box) {
	GList *kids = gtk_container_get_children((GtkContainer *)box);
	int i = 0, found = -1;
	for (GList *l = kids; l; l = l->next, i++)
		if (gtk_toggle_button_get_active(l->data)) {
			found = i;
			break;
		}
	g_list_free(kids);
	return found;
}

/* ---------- Картинка (Image) ----------
   Закодированные байты хранятся у виджета: при смене масштаба экрана (scale-factor)
   картинка перерисовывается из них и остаётся чёткой. src — только что декодированная
   картинка между load и show, чтобы не декодировать её дважды. */
typedef struct {
	guchar *data;
	gsize len;
	int w, h; /* размер показа, логические пиксели */
	GdkPixbuf *src;
} NuiImage;

static void nui_image_free(gpointer p) {
	NuiImage *im = p;
	g_free(im->data);
	if (im->src)
		g_object_unref(im->src);
	g_free(im);
}

/* Декодировать и повернуть по EXIF; 0 — не картинка. Новая ссылка. */
static GdkPixbuf *nui_decode(const guchar *data, gsize len) {
	GdkPixbufLoader *l = gdk_pixbuf_loader_new();
	int ok = gdk_pixbuf_loader_write(l, data, len, 0);
	ok = gdk_pixbuf_loader_close(l, 0) && ok; /* закрыть нужно и после ошибки */
	GdkPixbuf *pb = ok ? gdk_pixbuf_loader_get_pixbuf(l) : 0;
	if (pb)
		pb = gdk_pixbuf_apply_embedded_orientation(pb);
	g_object_unref(l);
	return pb;
}

/* Картинка размером w×h логических пикселей: w·k×h·k точек при масштабе k. */
static void nui_image_render(GtkWidget *w) {
	NuiImage *im = g_object_get_data((GObject *)w, "nui-image");
	if (!im->data || im->w <= 0)
		return;
	GdkPixbuf *src = im->src ? im->src : nui_decode(im->data, im->len);
	im->src = 0;
	if (!src)
		return;
	int k = gtk_widget_get_scale_factor(w), pw = im->w * k, ph = im->h * k;
	GdkPixbuf *pb = gdk_pixbuf_get_width(src) == pw && gdk_pixbuf_get_height(src) == ph
		? g_object_ref(src) : gdk_pixbuf_scale_simple(src, pw, ph, NUI_GDK_INTERP_BILINEAR);
	g_object_unref(src);
	if (!pb)
		return;
	cairo_surface_t *s = gdk_cairo_surface_create_from_pixbuf(pb, k, 0);
	g_object_unref(pb);
	gtk_image_set_from_surface((GtkImage *)w, s);
	cairo_surface_destroy(s);
}

static void nui_image_rescale(GtkWidget *w, gpointer pspec, gpointer data) {
	(void)pspec;
	(void)data;
	nui_image_render(w);
}

GtkWidget *nui_image_new(void) {
	GtkWidget *w = gtk_image_new();
	g_object_set_data_full((GObject *)w, "nui-image", g_malloc0(sizeof(NuiImage)), nui_image_free);
	g_signal_connect_data(w, "notify::scale-factor", (GCallback)nui_image_rescale, 0, 0, 0);
	return w;
}

/* len == 0 — убрать картинку. Иначе 0 — не картинка (прежняя остаётся), 1 — wh = размер. */
int nui_image_load(GtkWidget *w, const guchar *data, gsize len, int wh[2]) {
	NuiImage *im = g_object_get_data((GObject *)w, "nui-image");
	GdkPixbuf *pb = len ? nui_decode(data, len) : 0;
	if (len && !pb)
		return 0;
	g_free(im->data);
	im->data = 0;
	im->len = 0;
	if (im->src)
		g_object_unref(im->src);
	im->src = pb;
	if (!pb) {
		im->w = im->h = 0;
		gtk_image_clear((GtkImage *)w);
		return 1;
	}
	im->data = g_malloc(len);
	memcpy(im->data, data, len);
	im->len = len;
	wh[0] = gdk_pixbuf_get_width(pb);
	wh[1] = gdk_pixbuf_get_height(pb);
	return 1;
}

void nui_image_show(GtkWidget *w, int width, int height) {
	NuiImage *im = g_object_get_data((GObject *)w, "nui-image");
	im->w = width;
	im->h = height;
	nui_image_render(w);
}

/* ---------- List: виртуальная модель GtkTreeModel ----------
   Строки не хранятся: строка — номер в iter->user_data, текст ячейки при отрисовке
   спрашивается у Go (nuiListCell). Новые данные — новая модель (nui_list_set_rows);
   при её уничтожении Go забывает функцию ячеек (nuiListDrop). */
typedef struct {
	GObject parent;
	int rows, cols, stamp;
	uintptr_t key;
} NuiListModel;

static gpointer list_parent_class;

#define LM(m) ((NuiListModel *)(m))
#define ROW(it) ((int)(uintptr_t)(it)->user_data)

static gboolean lm_iter(NuiListModel *m, GtkTreeIter *it, int row) {
	if (row < 0 || row >= m->rows) {
		it->stamp = 0;
		return 0;
	}
	it->stamp = m->stamp;
	it->user_data = (gpointer)(uintptr_t)row;
	return 1;
}

static GtkTreeModelFlags lm_flags(GtkTreeModel *m) {
	(void)m;
	return NUI_GTK_TREE_MODEL_LIST_ONLY | NUI_GTK_TREE_MODEL_ITERS_PERSIST;
}
static gint lm_n_columns(GtkTreeModel *m) { return LM(m)->cols; }
static GType lm_column_type(GtkTreeModel *m, gint i) {
	(void)m;
	(void)i;
	return NUI_G_TYPE_STRING;
}
static gboolean lm_get_iter(GtkTreeModel *m, GtkTreeIter *it, GtkTreePath *p) {
	return gtk_tree_path_get_depth(p) == 1 && lm_iter(LM(m), it, gtk_tree_path_get_indices(p)[0]);
}
static GtkTreePath *lm_get_path(GtkTreeModel *m, GtkTreeIter *it) {
	(void)m;
	return gtk_tree_path_new_from_indices(ROW(it), -1);
}
static void lm_get_value(GtkTreeModel *m, GtkTreeIter *it, gint col, GValue *v) {
	g_value_init(v, NUI_G_TYPE_STRING);
	g_value_take_string(v, nuiListCell(LM(m)->key, ROW(it), col));
}
static gboolean lm_next(GtkTreeModel *m, GtkTreeIter *it) { return lm_iter(LM(m), it, ROW(it) + 1); }
static gboolean lm_previous(GtkTreeModel *m, GtkTreeIter *it) { return lm_iter(LM(m), it, ROW(it) - 1); }
static gboolean lm_children(GtkTreeModel *m, GtkTreeIter *it, GtkTreeIter *parent) {
	return !parent && lm_iter(LM(m), it, 0);
}
static gboolean lm_has_child(GtkTreeModel *m, GtkTreeIter *it) {
	(void)m;
	(void)it;
	return 0;
}
static gint lm_n_children(GtkTreeModel *m, GtkTreeIter *it) { return it ? 0 : LM(m)->rows; }
static gboolean lm_nth_child(GtkTreeModel *m, GtkTreeIter *it, GtkTreeIter *parent, gint n) {
	return !parent && lm_iter(LM(m), it, n);
}
static gboolean lm_parent(GtkTreeModel *m, GtkTreeIter *it, GtkTreeIter *child) {
	(void)m;
	(void)it;
	(void)child;
	return 0;
}

static void lm_iface_init(gpointer g_iface, gpointer data) {
	(void)data;
	GtkTreeModelIface *i = g_iface;
	i->get_flags = lm_flags;
	i->get_n_columns = lm_n_columns;
	i->get_column_type = lm_column_type;
	i->get_iter = lm_get_iter;
	i->get_path = lm_get_path;
	i->get_value = lm_get_value;
	i->iter_next = lm_next;
	i->iter_previous = lm_previous;
	i->iter_children = lm_children;
	i->iter_has_child = lm_has_child;
	i->iter_n_children = lm_n_children;
	i->iter_nth_child = lm_nth_child;
	i->iter_parent = lm_parent;
}

static void lm_finalize(GObject *o) {
	nuiListDrop(LM(o)->key);
	((GObjectClass *)list_parent_class)->finalize(o);
}

static void lm_class_init(gpointer klass, gpointer data) {
	(void)data;
	list_parent_class = g_type_class_peek_parent(klass);
	((GObjectClass *)klass)->finalize = lm_finalize;
}

static GType lm_type(void) {
	static GType t;
	if (!t) {
		t = g_type_register_static_simple(NUI_G_TYPE_OBJECT, "NuiListModel", sizeof(GObjectClass), lm_class_init,
			sizeof(NuiListModel), 0, 0);
		GInterfaceInfo info = {lm_iface_init, 0, 0};
		g_type_add_interface_static(t, gtk_tree_model_get_type(), &info);
	}
	return t;
}

/* row-activated (двойной щелчок, Enter): у сигнала лишние аргументы — свой обработчик. */
static void list_activated(GtkTreeView *v, GtkTreePath *p, GtkTreeViewColumn *c, gpointer ud) {
	(void)p;
	(void)c;
	nuiSignal(v, (uintptr_t)ud);
}

/* Список в своей прокрутке, как TextView. Высота строк одинаковая (fixed-height-mode):
   GTK не измеряет все строки, рисуются только видимые — миллион строк не тормозит. */
GtkWidget *nui_list_new(GtkWidget **view, uintptr_t activate_ud) {
	GtkWidget *sw = gtk_scrolled_window_new(0, 0);
	gtk_scrolled_window_set_policy((GtkScrolledWindow *)sw, NUI_GTK_POLICY_AUTOMATIC, NUI_GTK_POLICY_AUTOMATIC);
	GtkWidget *v = gtk_tree_view_new();
	GtkTreeView *tv = (GtkTreeView *)v;
	gtk_tree_view_set_fixed_height_mode(tv, 1);
	gtk_tree_view_set_enable_search(tv, 0);
	gtk_tree_selection_set_mode(gtk_tree_view_get_selection(tv), NUI_GTK_SELECTION_SINGLE);
	g_signal_connect_data(v, "row-activated", (GCallback)list_activated, (gpointer)activate_ud, 0, 0);
	gtk_container_add((GtkContainer *)sw, v);
	gtk_widget_show(v);
	*view = v;
	return sw;
}

/* Колонка с текстом колонки index модели; длинный текст — с многоточием. Последняя
   добавленная колонка забирает свободное место. Заголовки видны, если хоть один не пуст. */
void nui_list_column(GtkTreeView *v, const gchar *title, int index) {
	GtkCellRenderer *r = gtk_cell_renderer_text_new();
	g_object_set(r, "ellipsize", NUI_PANGO_ELLIPSIZE_END, (char *)0);
	GtkTreeViewColumn *c = gtk_tree_view_column_new();
	gtk_tree_view_column_set_title(c, title);
	gtk_tree_view_column_pack_start(c, r, 1);
	gtk_tree_view_column_add_attribute(c, r, "text", index);
	gtk_tree_view_column_set_sizing(c, NUI_GTK_TREE_VIEW_COLUMN_FIXED);
	gtk_tree_view_column_set_resizable(c, 1);
	if (index > 0)
		gtk_tree_view_column_set_expand(gtk_tree_view_get_column(v, index - 1), 0);
	gtk_tree_view_column_set_expand(c, 1);
	gtk_tree_view_append_column(v, c);
	gtk_tree_view_set_headers_visible(v, index > 0 || title[0]);
}

/* Ширина колонок — по заголовку и первым 100 строкам: все строки не измерить. */
static void list_autosize(GtkTreeView *v, NuiListModel *m) {
	PangoLayout *l = gtk_widget_create_pango_layout((GtkWidget *)v, "");
	for (int c = 0; c < m->cols; c++) {
		GtkTreeViewColumn *col = gtk_tree_view_get_column(v, c);
		int w = 0, h = 0, best = 40;
		const gchar *title = gtk_tree_view_column_get_title(col);
		if (title && title[0]) { /* заголовок — кнопка: поля по бокам */
			pango_layout_set_text(l, title, -1);
			pango_layout_get_pixel_size(l, &w, &h);
			best = w + 24 > best ? w + 24 : best;
		}
		for (int r = 0; r < m->rows && r < 100; r++) {
			gchar *s = nuiListCell(m->key, r, c);
			pango_layout_set_text(l, s, -1);
			g_free(s);
			pango_layout_get_pixel_size(l, &w, &h);
			if (w + 16 > best)
				best = w + 16;
		}
		gtk_tree_view_column_set_fixed_width(col, best < 400 ? best : 400);
	}
	g_object_unref(l);
}

/* Новые данные: новая модель (старую GTK отпустит — её функция ячеек забудется),
   выделение снимается, прокрутка — в начало. Курсор ставится на первую строку без
   выделения: иначе GtkTreeView в режиме SINGLE выделит её сам, как только получит
   фокус (gtk_tree_view_focus_to_cursor), а на Windows такого нет. */
void nui_list_set_rows(GtkTreeView *v, int rows, uintptr_t key) {
	NuiListModel *m = g_object_new(lm_type(), 0);
	m->rows = rows;
	m->cols = (int)gtk_tree_view_get_n_columns(v);
	m->stamp = (int)(key & 0x7fffffff) | 1;
	m->key = key;
	list_autosize(v, m);
	gtk_tree_view_set_model(v, (GtkTreeModel *)m);
	g_object_unref(m);
	if (rows > 0) {
		GtkTreePath *p = gtk_tree_path_new_from_indices(0, -1);
		gtk_tree_view_set_cursor(v, p, 0, 0); /* заодно прокрутка в начало */
		gtk_tree_path_free(p);
	}
	gtk_tree_selection_unselect_all(gtk_tree_view_get_selection(v));
}

int nui_list_selected(GtkTreeView *v) {
	GtkTreeIter it;
	if (!gtk_tree_selection_get_selected(gtk_tree_view_get_selection(v), 0, &it))
		return -1;
	return ROW(&it);
}

/* Выбрать строку, как это сделал бы пользователь: курсор, выделение, прокрутка к ней.
   row < 0 — снять выделение. */
void nui_list_select(GtkTreeView *v, int row) {
	GtkTreeModel *m = gtk_tree_view_get_model(v);
	if (row < 0 || !m || row >= LM(m)->rows) {
		gtk_tree_selection_unselect_all(gtk_tree_view_get_selection(v));
		return;
	}
	GtkTreePath *p = gtk_tree_path_new_from_indices(row, -1);
	gtk_tree_view_set_cursor(v, p, 0, 0);
	gtk_tree_path_free(p);
}

/* ---------- TextView: OnEnter и подсказка ---------- */

/* Enter без Shift — EvEnter вместо перевода строки (Shift+Enter — перевод строки). Сначала
   клавиша — методу ввода: Enter, который заканчивает набор (китайский, японский…), его. */
static gboolean text_key(GtkWidget *w, GdkEventKey *e, gpointer ud) {
	if (!g_object_get_data((GObject *)w, "nui-enter"))
		return 0;
	if ((e->keyval != 0xff0d && e->keyval != 0xff8d && e->keyval != 0xfe34) || (e->state & NUI_GDK_SHIFT_MASK))
		return 0;
	if (!gtk_text_view_im_context_filter_keypress((GtkTextView *)w, e))
		nuiSignal(w, (uintptr_t)ud);
	return 1;
}

void nui_text_enter(GtkTextView *v, int on, uintptr_t ud) {
	if (!g_object_get_data((GObject *)v, "nui-enter-connected")) {
		g_object_set_data_full((GObject *)v, "nui-enter-connected", (gpointer)1, 0);
		g_signal_connect_data(v, "key-press-event", (GCallback)text_key, (gpointer)ud, 0, 0);
	}
	g_object_set_data_full((GObject *)v, "nui-enter", (gpointer)(uintptr_t)(on != 0), 0);
}

/* Подсказка в пустом поле: рисуется после текста (draw, G_CONNECT_AFTER) с начала первой
   строки, цветом текста с половинной прозрачностью, как подсказка GtkEntry. */
static gboolean text_draw_hint(GtkWidget *w, cairo_t *cr, gpointer data) {
	(void)data;
	GtkTextView *v = (GtkTextView *)w;
	const gchar *s = g_object_get_data((GObject *)w, "nui-hint");
	GtkTextBuffer *b = gtk_text_view_get_buffer(v);
	if (!s || !s[0] || gtk_text_buffer_get_char_count(b) > 0)
		return 0;
	NuiTextIter it;
	int r[4], x, y;
	gtk_text_buffer_get_start_iter(b, (GtkTextIter *)&it);
	gtk_text_view_get_iter_location(v, (GtkTextIter *)&it, (GdkRectangle *)r);
	gtk_text_view_buffer_to_window_coords(v, NUI_GTK_TEXT_WINDOW_WIDGET, r[0], r[1], &x, &y);
	GdkRGBA c;
	gtk_style_context_get_color(gtk_widget_get_style_context(w), gtk_widget_get_state_flags(w), &c);
	PangoLayout *l = gtk_widget_create_pango_layout(w, s);
	int width = gtk_widget_get_allocated_width(w) - 2 * x;
	if (width > 0) {
		pango_layout_set_width(l, width * NUI_PANGO_SCALE);
		pango_layout_set_wrap(l, NUI_PANGO_WRAP_WORD_CHAR);
	}
	cairo_set_source_rgba(cr, c.red, c.green, c.blue, c.alpha * 0.5);
	cairo_move_to(cr, x, y);
	pango_cairo_show_layout(cr, l);
	g_object_unref(l);
	return 0;
}

static void text_hint_changed(GtkTextBuffer *b, gpointer view) {
	(void)b;
	gtk_widget_queue_draw(view); /* пустое ↔ непустое: подсказку стереть или нарисовать целиком */
}

void nui_text_placeholder(GtkTextView *v, const gchar *s) {
	if (!g_object_get_data((GObject *)v, "nui-hint-connected")) {
		g_object_set_data_full((GObject *)v, "nui-hint-connected", (gpointer)1, 0);
		g_signal_connect_data(v, "draw", (GCallback)text_draw_hint, 0, 0, 1 /* G_CONNECT_AFTER */);
		g_signal_connect_data(gtk_text_view_get_buffer(v), "changed", (GCallback)text_hint_changed, v, 0, 0);
	}
	g_object_set_data_full((GObject *)v, "nui-hint", g_strdup(s), g_free);
	gtk_widget_queue_draw((GtkWidget *)v);
}

/* ---------- OnDropFiles ---------- */

/* URI → локальные пути одной строкой «путь\0путь\0\0» (байты ФС как есть) — arg события
   EvDrop. Не локальные URI (http://…) пропускаются; не осталось путей — события нет. */
void nui_drop_uris(GtkWidget *w, gchar **uris, uintptr_t ud) {
	gsize n = 1;
	gchar *paths[256];
	int k = 0;
	for (int i = 0; uris && uris[i] && k < 256; i++)
		if ((paths[k] = g_filename_from_uri(uris[i], 0, 0)))
			n += strlen(paths[k++]) + 1;
	if (k == 0)
		return;
	gchar *buf = g_malloc(n), *p = buf;
	for (int i = 0; i < k; i++) {
		gsize l = strlen(paths[i]) + 1;
		memcpy(p, paths[i], l);
		p += l;
		g_free(paths[i]);
	}
	*p = 0;
	nuiSignalRet(w, buf, ud);
	g_free(buf);
}

/* drag-data-received: GTK_DEST_DEFAULT_ALL сам запрашивает данные и сам вызывает
   gtk_drag_finish. */
static void drop_received(GtkWidget *w, GdkDragContext *ctx, gint x, gint y, GtkSelectionData *sd, guint info,
	guint time, gpointer ud) {
	(void)ctx;
	(void)x;
	(void)y;
	(void)info;
	(void)time;
	gchar **uris = gtk_selection_data_get_uris(sd);
	nui_drop_uris(w, uris, (uintptr_t)ud);
	g_strfreev(uris);
}

void nui_drop_enable(GtkWidget *w, uintptr_t ud) {
	static GtkTargetEntry uriList = {"text/uri-list", 0, 0};
	if (g_object_get_data((GObject *)w, "nui-drop"))
		return;
	g_object_set_data_full((GObject *)w, "nui-drop", (gpointer)1, 0);
	gtk_drag_dest_set(w, NUI_GTK_DEST_DEFAULT_ALL, &uriList, 1, NUI_GDK_ACTION_COPY);
	g_signal_connect_data(w, "drag-data-received", (GCallback)drop_received, (gpointer)ud, 0, 0);
}

/* ---------- OnKey ---------- */
/* Латинская буква или цифра той же клавиши в любой раскладке: в русской раскладке
   keyval у Ctrl+S — «ы», а программе нужно 'S'. 0 — такой нет. */
guint nui_key_by_code(GdkEventKey *e) {
	GdkKeymap *km = gdk_keymap_get_for_display(gdk_display_get_default());
	guint *vals = 0, r = 0;
	gint n = 0;
	if (!km || !gdk_keymap_get_entries_for_keycode(km, e->hardware_keycode, 0, &vals, &n))
		return 0;
	for (int i = 0; i < n && !r; i++)
		if ((vals[i] >= 'a' && vals[i] <= 'z') || (vals[i] >= '0' && vals[i] <= '9'))
			r = vals[i];
	g_free(vals);
	return r;
}

/* Весь текст буфера. Строку освобождает вызывающий (g_free). GtkTextIter — только на стеке C. */
char *nui_text_get(GtkTextBuffer *b) {
	NuiTextIter s, e;
	gtk_text_buffer_get_start_iter(b, (GtkTextIter *)&s);
	gtk_text_buffer_get_end_iter(b, (GtkTextIter *)&e);
	return gtk_text_buffer_get_text(b, (GtkTextIter *)&s, (GtkTextIter *)&e, 0);
}

/* Дописать в конец; при max_lines > 0 удалить старые строки сверх max_lines; при autoscroll —
   прокрутить в конец. GTK считает пустую строку после завершающего \n отдельной строкой,
   поэтому она не учитывается. Прокрутка — через свою метку "nui-end": курсор и выделение
   пользователя не трогаются (place_cursor не используется). */
void nui_text_append(GtkTextView *v, const gchar *s, int len, int autoscroll, int max_lines) {
	GtkTextBuffer *b = gtk_text_view_get_buffer(v);
	NuiTextIter end;
	gtk_text_buffer_get_end_iter(b, (GtkTextIter *)&end);
	gtk_text_buffer_insert(b, (GtkTextIter *)&end, s, len);
	if (max_lines > 0) {
		gtk_text_buffer_get_end_iter(b, (GtkTextIter *)&end);
		int n = gtk_text_buffer_get_line_count(b) - (gtk_text_iter_starts_line((GtkTextIter *)&end) ? 1 : 0);
		if (n > max_lines) {
			NuiTextIter start, cut;
			gtk_text_buffer_get_start_iter(b, (GtkTextIter *)&start);
			gtk_text_buffer_get_iter_at_line(b, (GtkTextIter *)&cut, n - max_lines);
			gtk_text_buffer_delete(b, (GtkTextIter *)&start, (GtkTextIter *)&cut);
		}
	}
	if (autoscroll) {
		GtkTextMark *m = gtk_text_buffer_get_mark(b, "nui-end");
		gtk_text_buffer_get_end_iter(b, (GtkTextIter *)&end);
		if (m)
			gtk_text_buffer_move_mark(b, m, (GtkTextIter *)&end);
		else
			m = gtk_text_buffer_create_mark(b, "nui-end", (GtkTextIter *)&end, 0);
		gtk_text_view_scroll_mark_onscreen(v, m);
	}
}

/* Модальное сообщение: title — крупный заголовок, text — пояснение под ним.
   Внутри gtk_dialog_run крутится вложенный цикл: таймеры, Watch и Do продолжают работать.
   Своя ссылка на диалог: если родитель уничтожится, пока диалог открыт (DESTROY_WITH_PARENT),
   gtk_dialog_run вернётся, а объект ещё будет жив — destroy и unref безопасны. */
int nui_message(GtkWindow *parent, int type, int buttons, const gchar *title, const gchar *text) {
	GtkWidget *d = gtk_message_dialog_new(parent, NUI_GTK_DIALOG_MODAL | NUI_GTK_DIALOG_DESTROY_WITH_PARENT,
	                                      type, buttons, "%s", title);
	if (text && *text)
		gtk_message_dialog_format_secondary_text((GtkMessageDialog *)d, "%s", text);
	g_object_ref(d);
	int r = gtk_dialog_run((GtkDialog *)d);
	gtk_widget_destroy(d);
	g_object_unref(d);
	return r;
}

/* Ответ файлового менеджера на ShowItems: при ошибке (службы нет и т. п.) — открыть папку. */
static void nui_show_items_done(GObject *src, GAsyncResult *res, gpointer dir_uri) {
	GError *err = 0;
	GVariant *v = g_dbus_connection_call_finish((GDBusConnection *)src, res, &err);
	if (v) {
		g_variant_unref(v);
	} else {
		g_error_free(err);
		gtk_show_uri_on_window(0, (const char *)dir_uri, 0, 0);
	}
	g_free(dir_uri);
	g_object_unref(src);
}

/* Выделить файл в файловом менеджере: D-Bus org.freedesktop.FileManager1.ShowItems, асинхронно —
   интерфейс не ждёт. Таймаут D-Bus по умолчанию: при коротком файловый менеджер, запускавшийся
   дольше, всё равно выполнил бы ShowItems, и вместе с запасным путём открылись бы два окна.
   Нет сессии D-Bus — сразу открыть папку. 1 — запрос отправлен или папка открыта. */
int nui_show_items(const gchar *uri, const gchar *dir_uri) {
	GDBusConnection *c = g_bus_get_sync(NUI_G_BUS_TYPE_SESSION, 0, 0);
	if (!c)
		return gtk_show_uri_on_window(0, dir_uri, 0, 0);
	const gchar *uris[2] = {uri, 0};
	g_dbus_connection_call(c, "org.freedesktop.FileManager1", "/org/freedesktop/FileManager1",
	                       "org.freedesktop.FileManager1", "ShowItems", g_variant_new("(^ass)", uris, ""),
	                       0, NUI_G_DBUS_CALL_FLAGS_NONE, -1, 0, nui_show_items_done, g_strdup(dir_uri));
	return 1;
}

#ifdef NUI_GUI_TEST
/* Тестовый помощник: GtkAllocation живёт на стеке C, Go к нему не прикасается. */
void nui_allocation(GtkWidget *w, int out[4]) {
	GtkAllocation a;
	gtk_widget_get_allocation(w, &a);
	out[0] = a.x;
	out[1] = a.y;
	out[2] = a.width;
	out[3] = a.height;
}

/* Минимальный и естественный размер: ширина (out[0], out[1]), высота (out[2], out[3]). */
void nui_preferred(GtkWidget *w, int out[4]) {
	gtk_widget_get_preferred_width(w, &out[0], &out[1]);
	gtk_widget_get_preferred_height(w, &out[2], &out[3]);
}

/* Положение виджета в его окне верхнего уровня — сквозь GtkViewport, где allocation
   отсчитывается от собственного GdkWindow, — и начало этого окна на экране. */
void nui_window_pos(GtkWidget *w, int out[4]) {
	GtkWidget *top = gtk_widget_get_toplevel(w);
	GdkWindow *win = gtk_widget_get_window(top);
	out[0] = out[1] = out[2] = out[3] = 0;
	gtk_widget_translate_coordinates(w, top, 0, 0, &out[0], &out[1]);
	if (win)
		gdk_window_get_origin(win, &out[2], &out[3]);
}

/* Видимый диалог среди окон верхнего уровня: chooser — файловый (GtkFileChooserNative
   без портала показывает GtkFileChooserDialog), иначе — любой другой GtkDialog. */
static GtkWidget *nui_find_dialog(int chooser) {
	GList *all = gtk_window_list_toplevels(), *it;
	GtkWidget *found = 0;
	for (it = all; it; it = it->next) {
		GtkWidget *w = it->data;
		if (!gtk_widget_get_visible(w) || !g_type_check_instance_is_a((GTypeInstance *)w, gtk_dialog_get_type()))
			continue;
		if (!!g_type_check_instance_is_a((GTypeInstance *)w, gtk_file_chooser_get_type()) == !!chooser)
			found = w;
	}
	g_list_free(all);
	return found;
}

/* Открыт ли сейчас диалог (ничего в нём не нажимая). */
int nui_dialog_open(int chooser) { return nui_find_dialog(chooser) != 0; }

/* Ответить в открытом диалоге, как будто нажали кнопку. 0 — диалога нет. */
int nui_sim_dialog_response(int chooser, int response) {
	GtkWidget *d = nui_find_dialog(chooser);
	if (!d)
		return 0;
	gtk_dialog_response((GtkDialog *)d, response);
	return 1;
}

/* Выбрать путь в открытом файловом диалоге (каталог загружается асинхронно — дайте ему время). */
int nui_sim_file_chooser_set(const gchar *path) {
	GtkWidget *d = nui_find_dialog(1);
	if (!d)
		return 0;
	return gtk_file_chooser_set_filename((GtkFileChooser *)d, path);
}

/* Список файлов диалога — видимый GtkTreeView (боковая панель мест в 3.24 — GtkListBox). */
static void nui_find_tree(GtkWidget *w, gpointer out) {
	if (*(GtkWidget **)out)
		return;
	if (g_type_check_instance_is_a((GTypeInstance *)w, gtk_tree_view_get_type()) && gtk_widget_get_mapped(w))
		*(GtkWidget **)out = w;
	else if (g_type_check_instance_is_a((GTypeInstance *)w, gtk_container_get_type()))
		gtk_container_forall((GtkContainer *)w, nui_find_tree, out);
}

/* Текст ячейки так, как его видит GtkTreeView: через модель (gtk_tree_model_get). */
char *nui_list_cell_text(GtkTreeView *v, int row, int col) {
	GtkTreeModel *m = gtk_tree_view_get_model(v);
	GtkTreeIter it;
	gchar *s = 0;
	if (m && gtk_tree_model_iter_nth_child(m, &it, 0, row))
		gtk_tree_model_get(m, &it, col, &s, -1);
	return s;
}

/* Нажатие клавиши, как с клавиатуры: событие идёт окну верхнего уровня виджета w, а оно
   передаёт его виджету с фокусом, если обработчики окна не забрали. 1 — клавишу
   обработал кто-то: окно или виджет. code_of — keyval, по которому ищется клавиша
   (hardware_keycode): так, keyval «ы» с клавишей S — это Ctrl+S в русской раскладке. */
int nui_sim_key(GtkWidget *w, guint keyval, guint code_of, guint state) {
	GdkDisplay *d = gdk_display_get_default();
	GdkEventKey *e = (GdkEventKey *)gdk_event_new(NUI_GDK_KEY_PRESS);
	GdkWindow *win = gtk_widget_get_window(w);
	GdkKeymapKey *keys = 0;
	gint n = 0;
	e->window = win ? g_object_ref(win) : 0;
	e->send_event = 1;
	e->state = state;
	e->keyval = keyval;
	e->string = g_strdup("");
	if (gdk_keymap_get_entries_for_keyval(gdk_keymap_get_for_display(d), code_of, &keys, &n) && n > 0)
		e->hardware_keycode = (guint16)keys[0].keycode;
	g_free(keys);
	gdk_event_set_device((GdkEvent *)e, gdk_seat_get_keyboard(gdk_display_get_default_seat(d)));
	int r = gtk_widget_event(w, (GdkEvent *)e);
	gdk_event_free((GdkEvent *)e);
	return r;
}

/* Цвет точки (x, y) виджета на экране: 0xRRGGBB. */
unsigned nui_pixel_at(GtkWidget *w, int x, int y) {
	GdkPixbuf *pb = gdk_pixbuf_get_from_window(gtk_widget_get_window(w), x, y, 1, 1);
	if (!pb)
		return 0xFFFFFFFF;
	guchar *p = gdk_pixbuf_get_pixels(pb);
	unsigned c = (unsigned)p[0] << 16 | (unsigned)p[1] << 8 | p[2];
	g_object_unref(pb);
	return c;
}

/* Событие мыши, как от GDK: type — GDK_BUTTON_PRESS, RELEASE, MOTION_NOTIFY или SCROLL
   (для него button — направление). Идёт виджету через gtk_widget_event. */
void nui_sim_mouse(GtkWidget *w, int type, double x, double y, unsigned button, unsigned state) {
	GdkEvent *e = gdk_event_new(type);
	GdkWindow *win = gtk_widget_get_window(w);
	GdkEventButton *b = (GdkEventButton *)e;
	b->window = win ? g_object_ref(win) : 0;
	b->send_event = 1;
	if (type == NUI_GDK_MOTION_NOTIFY) {
		GdkEventMotion *m = (GdkEventMotion *)e;
		m->x = x, m->y = y, m->state = state;
	} else if (type == NUI_GDK_SCROLL) {
		GdkEventScroll *s = (GdkEventScroll *)e;
		s->x = x, s->y = y, s->state = state, s->direction = (int)button;
	} else {
		b->x = x, b->y = y, b->state = state, b->button = button;
	}
	gdk_event_set_device(e, gdk_seat_get_pointer(gdk_display_get_default_seat(gdk_display_get_default())));
	gtk_widget_event(w, e);
	gdk_event_free(e);
}

/* Двойной щелчок или Enter по строке. */
void nui_sim_list_activate(GtkTreeView *v, int row) {
	GtkTreePath *p = gtk_tree_path_new_from_indices(row, -1);
	gtk_tree_view_row_activated(v, p, gtk_tree_view_get_column(v, 0));
	gtk_tree_path_free(p);
}

/* Добавить файл к выбранным, как Ctrl+щелчок; каталог уже загружен. select_filename ставит
   на файл курсор и тем снимает прежнее выделение (show_and_select_files в GTK 3.24),
   поэтому выделенные строки запоминаются и выделяются снова. */
int nui_sim_file_chooser_add(const gchar *path) {
	GtkWidget *d = nui_find_dialog(1), *tv = 0;
	if (!d)
		return 0;
	nui_find_tree(d, &tv);
	if (!tv)
		return 0;
	GtkTreeSelection *sel = gtk_tree_view_get_selection((GtkTreeView *)tv);
	GList *rows = gtk_tree_selection_get_selected_rows(sel, 0), *it;
	int ok = gtk_file_chooser_select_filename((GtkFileChooser *)d, path);
	for (it = rows; it; it = it->next) {
		gtk_tree_selection_select_path(sel, it->data);
		gtk_tree_path_free(it->data);
	}
	g_list_free(rows);
	return ok;
}
#endif
