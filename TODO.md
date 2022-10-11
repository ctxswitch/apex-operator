# TODO

- [x] Default the scraper CRD
- [x] Create the metric interface
- [x] Create the prometheus input
- [x] Create null output
- [ ] Create logging output for testing
- [ ] Handle the reconciliation observer/desired state
- [ ] Manage the scraping process and start/stop the process on create/delete
- [ ] Manage the scraping process on update
- [ ] Wire up webhook defaults
- [ ] Handle webhook validations
- [x] Set up limited use datadog account
- [ ] Create datadog output
- [ ] Allow common tags from k8s information (namespace, pod, node, etc..)
- [ ] Allow annotations and labels for tags
- [ ] Support protobuf prometheus scraping
