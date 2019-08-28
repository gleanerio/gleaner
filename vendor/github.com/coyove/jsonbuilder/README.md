# Yet Another JSON Builder

## A quick look

```golang
type Foo struct {
	Name    string
	Age     int
	Address string
}

func main() {
	foo := Foo{"someone", 25, "earth"}
	tmp := jsonbuilder.From(foo)

	tmp.
		Begin("family").
		/*   */ Set("father", "his_father").
		/*   */ Set("mother", "his_mother").
		/*   */ Set("brothers", jsonbuilder.Array(
		/*   */ /*   */ "brother_1",
		/*   */ /*   */ "brother_2",
		/*   */ /*   */ "brother_3")).
		End()

	tmp.Enter("family").Set("sisters", "sister_1", "sister_2", "sister_3")
	tmp.Dive("family", "brothers").Set(2, jsonbuilder.Array("cousin_1", "cousin_2"))
	tmp.Enter("family").Enter("brothers").Enter(2).Delete(0).Leave().Delete(1)

	fmt.Println(tmp.MarshalPretty())
}

```
will give you:
```
{
    "Address": "earth",
    "Age": 25,
    "Name": "someone",
    "family": {
        "brothers": [
            "brother_1",
            [
                "cousin_2"
            ]
        ],
        "father": "his_father",
        "mother": "his_mother",
        "sisters": [
            "sister_1",
            "sister_2",
            "sister_3"
        ]
    }
}
```
## Usage
The code is pretty self-explanatory but there are 3 things to note:

1. `Enter(n)` and `Begin(n)` are the same, so are `Leave()` and `End()`. They are for writing better readable codes. If you tend to enter many levels, use `Dive(n...)`.
2. `Set(k, v...)` accepts multiple arguments. The 1st one is the key or index based on the type of `k` (`string` or `int`) and the rest are values. if there are multiple values, they will automatically form an array.
3. `Delete(n)` deletes a key from an object or an element from an array based on the type of `n` (`string` or `int`).