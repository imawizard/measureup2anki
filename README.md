# measureup2anki

is just a small tool to help importing your bought MeasureUp tests into Anki.

## Usage

1. Log into your MeasureUp account
2. Open the learning locker
3. Get your `PHPSESSID` e.g. from the Developer Tools' network tab
4. Clone the repository and run the tool with:

```sh
git clone https://github.com/imawizard/measureup2anki measureup2anki && cd $_
go run . dump $COOKIE [$TEST] # Insert your PHPSESSID here
                              # Leave 2nd param empty to get a list instead
go run . produce [$TEST]
```

5. Create a deck in Anki and set up the card type
6. Import the .csv in `/out` into the Anki deck

## Anki Card

### Front Template

```html
{{Text}}

{{#Image}}
	<br>
	<br>
	{{Image}}
{{/Image}}

<ul>
	<div id="option-1" class="option front hidden"></div>
	<div id="option-2" class="option front hidden"></div>
	<div id="option-3" class="option front hidden"></div>
	<div id="option-4" class="option front hidden"></div>
	<div id="option-5" class="option front hidden"></div>
	<div id="option-6" class="option front hidden"></div>
	<div id="option-7" class="option front hidden"></div>
	<div id="option-8" class="option front hidden"></div>
</ul>

{{#Exhibits}}
	<hr id="exhibits">
	{{Exhibits}}
{{/Exhibits}}

<script>

var get = (i) => document.getElementById("option-" + i);
var show = (el) => el.classList.remove("hidden");

var options = [
	`{{Option-1}}`,
	`{{Option-2}}`,
	`{{Option-3}}`,
	`{{Option-4}}`,
	`{{Option-5}}`,
	`{{Option-6}}`,
	`{{Option-7}}`,
	`{{Option-8}}`,
].filter(o => String(o));

options
	.map((o, i) => ["{{Type}}" !== "liveScreen" ? Math.random() * options : i.length, o])
	.sort()
	.map(p => p[1])
	.forEach((o, i) => {
		var el = get(i + 1);
		show(el);
		el.innerHTML = "<li>" + o + "</li>";
	});

</script>
```

### Back Template

```html
{{Image}}

<ul>
	<div id="option-1" class="option back hidden"><li>{{Option-1}}</li></div>
	<div id="option-2" class="option back hidden"><li>{{Option-2}}</li></div>
	<div id="option-3" class="option back hidden"><li>{{Option-3}}</li></div>
	<div id="option-4" class="option back hidden"><li>{{Option-4}}</li></div>
	<div id="option-5" class="option back hidden"><li>{{Option-5}}</li></div>
	<div id="option-6" class="option back hidden"><li>{{Option-6}}</li></div>
	<div id="option-7" class="option back hidden"><li>{{Option-7}}</li></div>
	<div id="option-8" class="option back hidden"><li>{{Option-8}}</li></div>
</ul>

<hr id="explanation">

{{Explanation}}

<script>

var get = (i) => document.getElementById("option-" + i);
var getAll = () => Array.of(...document.getElementsByClassName("option"));
var show = (el) => el.classList.remove("hidden");
var wrong = (el) => el.classList.add("wrong");

var answer = {{Answer}};

switch ("{{Type}}") {
case "singleChoice":
  show(get(answer));
	break;
case "multipleChoice":
	answer.forEach((i) => show(get(i)));
	break;
case "contentTable":
	getAll().forEach((el, i) => {
		if (el.textContent == "") {
			return;
		}
		show(el);
		if (!answer.includes(i)) {
			wrong(el);
		}
	});
	break;
case "liveScreen":
	for (var i = 1; i <= answer.length; i++) {
		var el = get(i);
		show(el);
		var il = el.children[0];
		il.innerHTML = il.innerHTML.split(" ╱ ")[answer[i - 1]];
	}
	break;
case "buildList":
case "buildListReorder":
	// Do nothing as the explanation already includes the answer.
case "selectPlaceMup":
	// Do nothing, just show the explanation.
	break;
}

</script>
```

### Styling

```css
.card {
  font-family: arial;
  font-size: 18px;
  //text-align: center;
  color: black;
  background-color: white;
}

.option {
  list-style-type: circle;
}

.image {
  border: solid 1px;
}

.hidden {
  display: none;
}

.wrong {
  text-decoration: line-through;
}
```
