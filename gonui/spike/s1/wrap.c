/* wrap.c — то, что Go не может вызвать сам: макрос g_signal_connect,
   функции с переменным числом аргументов, указатели на колбэки. */
#include "s1.h"

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

guint nui_watch(int fd, unsigned cond, uintptr_t ud) {
	cond |= NUI_G_IO_HUP | NUI_G_IO_ERR | NUI_G_IO_NVAL;
	return g_unix_fd_add_full(NUI_G_PRIORITY_DEFAULT, fd, (GIOCondition)cond,
	                          (GUnixFDSourceFunc)nuiWatch, (gpointer)ud, 0);
}

static GtkWidget *cur_dialog; /* открытый диалог — для автопилота */

/* Вопрос «Да/Нет». Внутри gtk_dialog_run крутится вложенный цикл: таймеры и Do продолжают работать. */
int nui_ask(GtkWindow *parent, const gchar *text) {
	GtkWidget *d = gtk_message_dialog_new(parent, NUI_GTK_DIALOG_MODAL | NUI_GTK_DIALOG_DESTROY_WITH_PARENT,
	                                      NUI_GTK_MESSAGE_QUESTION, NUI_GTK_BUTTONS_YES_NO, "%s", text);
	cur_dialog = d;
	int r = gtk_dialog_run((GtkDialog *)d);
	cur_dialog = 0;
	gtk_widget_destroy(d);
	return r == NUI_GTK_RESPONSE_YES;
}

/* Только для автопилота: ответить в открытом диалоге, как будто нажали кнопку. */
void nui_test_respond(int yes) {
	if (cur_dialog)
		gtk_dialog_response((GtkDialog *)cur_dialog, yes ? NUI_GTK_RESPONSE_YES : NUI_GTK_RESPONSE_NO);
}
