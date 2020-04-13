window.onload = function () {
    document.querySelector('input').addEventListener('keydown', function (e) {
        if (e.keyCode === 13) {
            e.preventDefault();
            send(this.value);
            console.log(this.value);
            this.value = "";
        }
    });
    const request = new XMLHttpRequest();

    function send(userInput) {
        function s() {

            if (request.readyState === 4) {
                const status = request.status;
                if (status === 200) {
                    jsonParse(request.responseText);
                    console.log(request.responseText);
                    request.abort();
                }
            }
        }

        request.open("GET", "http://localhost:8888/api?search=" + userInput);
        request.onreadystatechange = s;
        request.send();

    }

    function jsonParse(json) {
        let res = JSON.parse(json);
        // document.getElementById("result").innerHTML = createTemplate(res[0].filename, res[0].wordsEncountered);
        document.getElementById("hidden-block").innerHTML = createTemplate(res[0].filename, res[0].wordsEncountered);
        console.log(res[0].wordsEncountered);
    }

    function createTemplate(filename, words) {
        return "<span class='result' style='background: white;  margin-bottom: 20px;  display: block;  margin-top: 5px;  width: 450px;  height: auto;  padding: 10px 0;  border: 1px solid #eee;  border-radius: 20px;'> " +
        " <span class='file' style='display: block; margin: 5px 20px;' > " +
            " <span class='title-file' style='display: inline-block' >" + filename + " </span> " +
            "<span class='words-encountered' style='display: inline-block; float: right; padding-left: 20px; border-left: 1px solid #eee;' >" + words + "</span>" +
            "</span> " +
        "</span> "
    }
};