var searchBox = document.getElementById('search-box');
var apps = document.querySelectorAll('li[data-app]');

function search() {
  var node, name;
  for (var i = 0; i < apps.length; i++) {
    node = apps[i];
    name = node.attributes['data-app'].value;
    node.classList.toggle('hide', name.search(searchBox.value) < 0);
  }
}

search();
