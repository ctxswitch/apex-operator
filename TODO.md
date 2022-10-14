# TODO

- [x] Default the scraper CRD
- [x] Create the metric interface
- [x] Create the prometheus input
- [x] Create null output
- [x] Add prometheus input configuration to CRD and input
- [x] Add ouptut specification for null output  to CRD
- [ ] Handle the reconciliation observer/desired state
- [ ] Manage the scraping process and start/stop the process on create/delete
- [ ] Manage the scraping process on update
- [ ] Create logging output for testing
- [x] Set up limited use datadog account
- [ ] Create datadog output
- [ ] Allow common tags from k8s information (namespace, pod, node, etc..)
- [ ] Allow annotations and labels for tags
- [ ] Wire up webhook defaults
- [ ] Handle webhook validations
- [ ] Support protobuf prometheus scraping
