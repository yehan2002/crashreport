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

})