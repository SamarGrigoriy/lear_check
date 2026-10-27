/* rootshot — снимок экрана одним XGetImage(root), в PPM на stdout:
 *     rootshot | pnmtopng > shot.png
 *
 * Почему не xwd -root: GTK 3 создаёт окна с визуалом, пригодным для OpenGL, а не
 * с визуалом экрана по умолчанию. Видя на экране несколько визуалов, xwd склеивает
 * картинку по кусочкам и теряет перекрывающиеся окна: модальный диалог поверх
 * родителя выходит чёрным, хотя на экране он нарисован (проверено в спайке S1).
 *
 * Прототипы Xlib — свои, без заголовков (как в nui, ADR-3). Только TrueColor 24/32 бит. */
typedef struct _XDisplay Display;
typedef unsigned long Window;
typedef struct _XImage XImage;
Display *XOpenDisplay(const char *);
int XDefaultScreen(Display *);
Window XRootWindow(Display *, int);
int XDisplayWidth(Display *, int);
int XDisplayHeight(Display *, int);
XImage *XGetImage(Display *, unsigned long, int, int, unsigned, unsigned, unsigned long, int);
unsigned long XGetPixel(XImage *, int, int);
int printf(const char *, ...);
int putchar(int);

#define ZPIXMAP 2

int main(void) {
	Display *d = XOpenDisplay(0);
	if (!d)
		return 2;
	int s = XDefaultScreen(d), w = XDisplayWidth(d, s), h = XDisplayHeight(d, s);
	XImage *img = XGetImage(d, XRootWindow(d, s), 0, 0, w, h, ~0UL, ZPIXMAP);
	if (!img)
		return 3;
	printf("P6\n%d %d\n255\n", w, h);
	for (int y = 0; y < h; y++)
		for (int x = 0; x < w; x++) {
			unsigned long p = XGetPixel(img, x, y);
			putchar((p >> 16) & 255);
			putchar((p >> 8) & 255);
			putchar(p & 255);
		}
	return 0;
}
