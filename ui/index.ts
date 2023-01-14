import testData from "./data.json";
import dotT from "dot";


console.log(testData);
document.addEventListener("DOMContentLoaded", () => {
    let compilerInfo = `${testData.SysInfo.Compiler}`;
    let stackTraceTemplate = document.getElementById("stack-trace-template");
    if (!stackTraceTemplate) return;


    document.getElementById("sys-info").innerText = JSON.stringify(testData.SysInfo);
    document.getElementById("compiler-info").innerText = compilerInfo;

    var dependencies = new Map<string, typeof testData.Build.Deps[0]>();

    testData.Build.Deps.forEach((dep) => {
        dependencies.set(dep.Path, dep);
    });

    dependencies.set(testData.Build.Main.Path, testData.Build.Main);
    let template = dotT.template(document.getElementById("stack-trace-template")?.outerHTML);
    console.log(template)
    const parser = new DOMParser();
    (() => {
        testData.Stack.forEach((goroutine) => {
            const routine = parser.parseFromString(template(goroutine), "text/html").body.children[0];
            console.log(routine);
            goroutine.Stack.forEach((entry) => {
                var frame = document.createElement("div")
                frame.classList.add("px-3");
                let fn = entry.Func.split("/").reverse()[0];
                let pkg = fn.split(".")[0];
                let file = entry.File.split("/").reverse()[0];
                let fileLocation = `${file}:${entry.Line}`;

                let filePathParts = entry.File.split("/");
                for (let i = 0; i < filePathParts.length; i++) {
                    filePathParts.pop();
                    let filePathJoined = filePathParts.join("/")
                    dependencies.forEach((k, v) => {
                        if (filePathJoined.endsWith(k.Path)) {
                            console.log(v)
                        }
                    })
                };


                frame.innerText = `${pkg} ${fileLocation} ${fn}`;
                routine.appendChild(frame)
            })
            document.body.appendChild(routine);
        })
    })()

})