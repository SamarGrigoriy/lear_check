package drv

/*
#include "gtk3_linux.h"
*/
import "C"

import "nui/internal/core"

// MessageBox показывает модальное сообщение. true — нажата OK или «Да»; Esc — false.
func MessageBox(parent H, kind uint8, title, text string) bool {
	typ, buttons := C.int(C.NUI_GTK_MESSAGE_INFO), C.int(C.NUI_GTK_BUTTONS_OK)
	switch kind {
	case MsgAlert:
		typ = C.NUI_GTK_MESSAGE_ERROR
	case MsgAsk:
		typ, buttons = C.NUI_GTK_MESSAGE_QUESTION, C.NUI_GTK_BUTTONS_YES_NO
	}
	r := C.nui_message((*C.GtkWindow)(parent), typ, buttons, cstr(title), cstr(text))
	return r == C.NUI_GTK_RESPONSE_OK || r == C.NUI_GTK_RESPONSE_YES
}

// FileDialog — системный диалог выбора файла (GtkFileChooserNative: в песочнице — портал).
// Подписи кнопок — по умолчанию GTK, переведённые на язык системы. Save сам спрашивает
// о перезаписи; name — предлагаемое имя. Пути возвращаются как есть, в кодировке ФС;
// nil — отмена. Один путь, кроме FileOpenMany.
func FileDialog(parent H, mode uint8, title, name string, filters []Filter) []string {
	action := C.GtkFileChooserAction(C.NUI_GTK_FILE_CHOOSER_ACTION_OPEN)
	switch mode {
	case FileSave:
		action = C.NUI_GTK_FILE_CHOOSER_ACTION_SAVE
	case FileFolder:
		action = C.NUI_GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER
	}
	n := C.gtk_file_chooser_native_new(cstr(title), (*C.GtkWindow)(parent), action, nil, nil)
	nd := (*C.GtkNativeDialog)(C.gpointer(n))
	fc := (*C.GtkFileChooser)(C.gpointer(n))
	C.gtk_native_dialog_set_modal(nd, 1)
	if mode == FileOpenMany {
		C.gtk_file_chooser_set_select_multiple(fc, 1)
	}
	if mode == FileSave {
		C.gtk_file_chooser_set_do_overwrite_confirmation(fc, 1)
		if name != "" {
			C.gtk_file_chooser_set_current_name(fc, cstr(name))
		}
	}
	for _, f := range filters {
		ff := C.gtk_file_filter_new()
		C.gtk_file_filter_set_name(ff, cstr(f.Name))
		for _, p := range f.Patterns {
			C.gtk_file_filter_add_pattern(ff, cstr(p))
		}
		C.gtk_file_chooser_add_filter(fc, ff)
	}
	var paths []string
	if C.gtk_native_dialog_run(nd) == C.NUI_GTK_RESPONSE_ACCEPT {
		list := C.gtk_file_chooser_get_filenames(fc)
		for it := list; it != nil; it = it.next {
			paths = append(paths, gostrFree((*C.gchar)(it.data)))
		}
		C.g_slist_free(list)
	}
	C.g_object_unref(C.gpointer(n))
	return paths
}

// OpenURI открывает "https://…" или "file:///…" программой по умолчанию.
func OpenURI(parent H, uri string) bool {
	var err *C.GError
	ok := C.gtk_show_uri_on_window((*C.GtkWindow)(parent), (*C.char)(C.gpointer(cstr(uri))), 0, &err) != 0
	if err != nil {
		C.g_error_free(err)
	}
	return ok
}

// ShowInFolder просит файловый менеджер выделить файл (асинхронно), при ошибке — открывает
// папку. path — абсолютный путь в ФС; URI строит GLib, байты имени не меняются (cpath).
func ShowInFolder(path string) bool {
	uri := C.g_filename_to_uri(cpath(path), nil, nil)
	dir := C.g_filename_to_uri(cpath(core.Dir(path)), nil, nil)
	ok := false
	if uri != nil && dir != nil {
		ok = C.nui_show_items(uri, dir) != 0
	}
	if uri != nil {
		C.g_free(C.gpointer(uri))
	}
	if dir != nil {
		C.g_free(C.gpointer(dir))
	}
	return ok
}
