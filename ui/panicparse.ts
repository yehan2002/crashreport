import testData from "./data2.json";

const parser = new DOMParser()
const doc = parser.parseFromString(testData["data"], "text/html");

doc.head.querySelectorAll("meta").forEach(e => {
    if (e.name && e.name !== "viewport") {
        e.remove()
        document.head.append(e)
    }
});

const content = doc.querySelector("#content");
if (!content) {
    throw "unable to find content";
}

content.querySelectorAll("a").forEach(a => {
    if (a.hasAttribute("href")) {
        a.target = "_blank";
    }
})

document.body.appendChild(content);