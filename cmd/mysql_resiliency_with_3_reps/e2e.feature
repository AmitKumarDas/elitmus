Feature: Test MySQL resiliency on OpenEBS volume
  In order to test resiliency of MySQL on OpenEBS
  As an end user
  I need to be able to launch MySQL on OpenEBS
  I need to be able to use OpenEBS with replication factor of 3
  I need to have MySQL running even when 33% of volume nodes are un-available

  Scenario: launch MySQL on OpenEBS volume
    Given I have a kubernetes cluster with volume "operator" installed
    When I launch application "mysql-db" on volume "mysql-on-openebs"
    Then wait for "60s"
    And verify application is launched successfully on volume

  Scenario: OpenEBS volume replicas should run on unique nodes
    Given application is launched successfully on volume
    Then verify each volume replica gets a unique node
    And save "identifier" volume replica "replicaName" and node "nodeName"

  Scenario: MySQL application should run when first node is rebooted
    Given I shutdown "first" node
    Then wait for "60s"
    And verify "first" node is not available
    And verify application is running
    And verify "33%" volume replicas are running
    Then start "first" node
    And wait for "60s"
    And verify "first" node is running
    And verify application is running
    And verify "100%" volume replicas are running

  Scenario: MySQL application should run when second node is rebooted
    Given I shutdown "second" node
    Then wait for "60s"
    And verify "second" node is not available
    And verify application is running
    And verify "33%" volume replicas are running
    Then start "second" node
    And wait for "60s"
    And verify "second" node is running
    And verify application is running
    And verify "100%" volume replicas are running
