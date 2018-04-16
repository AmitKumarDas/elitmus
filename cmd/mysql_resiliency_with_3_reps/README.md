### Scenario: MySQL with 3 openebs replicas

#### Use-Case
Feature: MySQL on OpenEBS with replication factor of 3
  In order to test resiliency of MySQL on OpenEBS
  As an end user
  I need to be able to launch MySQL on OpenEBS
  I need to be able to use OpenEBS with replication factor of 3
  I need to have MySQL running even when 33% of volume nodes are un-available

#### Implementation
- Step 1: Describe the scenario(s) in scenario.feature file
- Step 2: Run **godog scenario.feature**
- Step 3: Implement undefined steps (also referred to as snippets) in scenario_test.go file
- Step 4: Re-Run **godog scenario.feature**

#### Best Practices
- 1: Make use of standard go practices
- 2: Transform the usecase into structure(s) & its properties
- 3: Now fit the godog generated function snippets into above structure' methods

#### References
- https://github.com/DATA-DOG/godog
- https://en.wikipedia.org/wiki/Literate_programming
