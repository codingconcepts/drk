# drk
[wrk](https://github.com/wg/wrk) but for databases

Cluster

```sh
cockroach demo --insecure --no-example-database
```

### Examples

Run the e-Commerce example

```sh
make ecommerce_example
```

Run the payments example

```sh
make payments_example
```

### Todos

* Fix all of the errors in the e-Commerce example before 50 checkouts
* Configure a workflow query for the exec type to test it
* Add the ability to ensure uniqueness across two arg values (re-running until unique, or crashing after X attempts)
* Update ref to allow more than one item to be seleted (e.g. add multiple products to a basket)
* TEST!
* Optionally pass args in workflow queries
* Ramp VU's up and down
