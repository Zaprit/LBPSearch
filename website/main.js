function setSearch() {
    const query = document.getElementById("query");
    const queryBg = document.getElementById("query_bg");
    const type = document.getElementById("type");

    let value = query.value;
    if (value.startsWith("user:")) {
        type.setAttribute("value","user");
        queryBg.setAttribute("value", value.split(":")[1]);
    } else {
        type.setAttribute("value","slot");
        queryBg.setAttribute("value", value);
    }
}


