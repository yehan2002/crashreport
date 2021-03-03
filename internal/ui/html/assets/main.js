const heartbeat = window.location.origin + "/ok";
class UI {
    constructor() {
        this.frame = document.getElementById("frame");
    }
    run() {
        this.awaitOnline().then(() => {
            this.openWebsocket();
            this.loadMainFrame();
        });
    }
    /**
     * awaitOnline returns a Promise that gets resolved when a successful connection is
     * made to the server.
     */
    awaitOnline() {
        return new Promise((resolve) => {
            let run = function () {
                fetch(heartbeat)
                    .then(() => {
                    if (this.shouldReload) {
                        window.location.reload();
                    }
                    resolve();
                })
                    .catch(() => {
                    setTimeout(run, 1000);
                });
            };
            run();
        });
    }
    loadMainFrame() {
        new Promise((resolve, reject) => {
            if (window.location.hash === "") {
                return reject();
            }
            var path = window.location.origin + decodeURIComponent(window.location.hash).replace("#", "");
            fetch(path).catch(() => {
                reject();
            }).then((w) => {
                if (w.ok) {
                    return resolve(path);
                }
                reject();
            });
        }).then((v) => {
            this.initFrame(v);
        }).catch(() => {
            let all = document.getElementsByClassName("dropdown-content")[0];
            let url = this.getElementOrDefault(all, all.children[0], "stacktrace", "Info");
            window.location.hash = "";
            this.initFrame(url.href);
        });
    }
    initFrame(url) {
        this.frame.src = url;
        this.frame.onload = () => {
            window.location.hash = encodeURIComponent(this.frame.contentDocument.location.pathname + this.frame.contentDocument.location.search);
            document.title = this.frame.contentDocument.title;
        };
    }
    openWebsocket() {
        let ws = new WebSocket("ws://" + window.location.host + "/websocket");
        ws.onclose = () => {
            this.run();
        };
        ws.onmessage = () => {
            window.location.reload();
        };
        ws.onopen = () => {
            this.shouldReload = true;
        };
    }
    /**
     * gets the element by the class name
     * @param parent the parent element
     * @param def the default element to return
     * @param classNames a array of class names for elements
     */
    getElementOrDefault(parent, def, ...classNames) {
        if (parent == null) {
            parent = document.body;
        }
        let result = def;
        classNames.reverse().forEach(className => {
            let elements = parent.getElementsByClassName(className);
            if (elements.length !== 0) {
                result = elements[0];
            }
        });
        return result;
    }
}
(function () {
    let ui = new UI();
    ui.run();
})();
