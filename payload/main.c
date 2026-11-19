/*
 * ps5-rpc payload
 *
 * Listens on a TCP port and answers every connection with ONE line of JSON
 * describing the foreground title, then closes the connection.
 *
 *   {"v":1,"running":true,"title_id":"PPSA01234","title_name":"Game"}\n
 *
 * Build with the ps5-payload-sdk (https://github.com/ps5-payload-dev/sdk):
 *   export PS5_PAYLOAD_SDK=/opt/ps5-payload-sdk && make
 *
 * STATUS: verified on a PS5 (fw 13.20, Relapse jailbreak).
 * Still untested: other firmware versions, games without a param.json at
 * the first path, and behaviour while suspended in rest mode.
 */
#include <arpa/inet.h>
#include <netinet/in.h>
#include <stdio.h>
#include <string.h>
#include <sys/socket.h>
#include <unistd.h>

#define PORT 9045

typedef struct {
    int  running;
    char title_id[16];    /* PPSA01234 */
    char title_name[128];
} title_info_t;

/*
 * The SDK ships only asm stubs for libSceSystemService (no headers), so the
 * prototypes are declared here.
 *
 *  - sceSystemServiceGetAppIdOfRunningBigApp(): signature taken from
 *    ps5-payload-dev/websrv (src/sys.h) which uses it in production.
 *  - sceSystemServiceGetAppTitleId(): NOT verified against a header or
 *    existing payload. Believed to be (int app_id, char *title_id_out).
 *    Symbol existence in the SDK stub is only partially confirmed.
 *    If it fails to link/resolve, the alternative is sceKernelGetAppInfo().
 */
extern int sceSystemServiceGetAppIdOfRunningBigApp(void);
extern int sceSystemServiceGetAppTitleId(int app_id, char *title_id_out);

/*
 * Fills `out` with the foreground game. Returns 0 on success, including the
 * "nothing running" case (out->running = 0). Needs on-console verification.
 */
static int get_foreground_title(title_info_t *out) {
    char tid[64] = {0};
    memset(out, 0, sizeof(*out));

    int app_id = sceSystemServiceGetAppIdOfRunningBigApp();
    if (app_id <= 0) return 0;                       /* home screen / no game */

    if (sceSystemServiceGetAppTitleId(app_id, tid) != 0 || !tid[0]) return 0;

    out->running = 1;
    snprintf(out->title_id, sizeof(out->title_id), "%s", tid);
    return 0;
}

/*
 * Resolve a human readable name from the title's param.json ("titleName").
 * Paths differ per firmware/storage; try the usual suspects.
 * Falls back to the title id (handled by the companion if name is empty).
 */
static void lookup_title_name(title_info_t *t) {
    static const char *fmts[] = {
        "/user/app/%s/sce_sys/param.json",   /* installed games (pattern seen in websrv) */
        "/user/appmeta/%s/param.json",
        "/system_data/priv/appmeta/%s/param.json",
        "/system_ex/app/%s/sce_sys/param.json",
    };
    char path[256], buf[8192];
    for (size_t i = 0; i < sizeof(fmts) / sizeof(*fmts); i++) {
        snprintf(path, sizeof(path), fmts[i], t->title_id);
        FILE *f = fopen(path, "r");
        if (!f) continue;
        size_t n = fread(buf, 1, sizeof(buf) - 1, f);
        fclose(f);
        buf[n] = 0;

        /* Naive scan: first "titleName":"..." -- good enough for param.json. */
        char *p = strstr(buf, "\"titleName\"");
        if (!p || !(p = strchr(p + 11, ':')) || !(p = strchr(p, '"'))) continue;
        p++;
        size_t o = 0;
        while (*p && *p != '"' && o < sizeof(t->title_name) - 1) {
            if (*p == '\\' && p[1]) p++;       /* drop escapes; Unicode left as UTF-8 */
            t->title_name[o++] = *p++;
        }
        t->title_name[o] = 0;
        if (o) return;
    }
}

/* Minimal JSON string escaping for the name field. */
static void json_escape(const char *in, char *out, size_t cap) {
    size_t o = 0;
    for (; *in && o + 7 < cap; in++) {
        unsigned char c = (unsigned char)*in;
        if (c == '"' || c == '\\') { out[o++] = '\\'; out[o++] = c; }
        else if (c < 0x20)         { o += snprintf(out + o, cap - o, "\\u%04x", c); }
        else                       { out[o++] = c; }
    }
    out[o] = 0;
}

int main(void) {
    int srv = socket(AF_INET, SOCK_STREAM, 0);
    if (srv < 0) return 1;

    int one = 1;
    setsockopt(srv, SOL_SOCKET, SO_REUSEADDR, &one, sizeof(one));

    struct sockaddr_in addr = {0};
    addr.sin_family      = AF_INET;
    addr.sin_addr.s_addr = htonl(INADDR_ANY);
    addr.sin_port        = htons(PORT);
    if (bind(srv, (struct sockaddr *)&addr, sizeof(addr)) < 0) return 2;
    if (listen(srv, 4) < 0) return 3;

    for (;;) {
        int c = accept(srv, NULL, NULL);
        if (c < 0) continue;

        title_info_t t;
        get_foreground_title(&t);
        if (t.running && t.title_id[0]) lookup_title_name(&t);

        char name[300], line[512];
        json_escape(t.title_name, name, sizeof(name));
        int n = snprintf(line, sizeof(line),
            "{\"v\":1,\"running\":%s,\"title_id\":\"%s\",\"title_name\":\"%s\"}\n",
            t.running ? "true" : "false", t.title_id, name);
        write(c, line, n);
        close(c);
    }
}
