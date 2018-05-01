# elasticsearch-query-test

Simple tool to check the accuracy of elasticsearch query results.

## Instructions

* Change the query endpoint in `main.go`
* Fill the `expectations.csv` file with test cases

  * `search_term`: input search term
  * `result_field`: source field that the result will be compared
  * `result_value`: expected result value
  * `result_position`: expected position of the result
  * `description`: (optional) description of the expected result

## Output:

<img src="./screenshot.png">
