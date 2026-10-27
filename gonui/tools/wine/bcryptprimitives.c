/* Заглушка bcryptprimitives.dll — только для запуска под Wine 8.0 (Debian 12), где этой
   библиотеки нет. Рантайм Go ≥ 1.22 берёт из неё ProcessPrng; на Windows 10+ она системная.
   В exe и в nui не попадает: tools/wine-run.sh кладёт её в префикс Wine. */
typedef int BOOL;
typedef unsigned char BYTE;
typedef unsigned long long SIZE_T;
typedef unsigned long ULONG;
BYTE __stdcall SystemFunction036(void *buf, ULONG len); /* RtlGenRandom, advapi32 */

BOOL __stdcall ProcessPrng(BYTE *data, SIZE_T len) {
	while (len > 0) {
		ULONG n = len > 0x40000000 ? 0x40000000 : (ULONG)len;
		if (!SystemFunction036(data, n))
			return 0;
		data += n;
		len -= n;
	}
	return 1;
}
