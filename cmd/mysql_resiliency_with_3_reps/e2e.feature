Feature: Test MySQL resiliency on Kubernetes PV
  In order to test resiliency of MySQL on Kubernetes PV
  As an end user
  I need to be able to launch MySQL on Kubernetes PV
  I need to be able to use PV with replication factor of 3
  I need to have MySQL running even when 33% of volume replicas are un-available

  Scenario: launch MySQL on Kubernetes PV
    Given I have a kubernetes cluster with volume operator installed
    When I launch mysql application on volume
    Then wait for "60s"
    And verify mysql application is launched successfully on volume

  Scenario: Kubernetes volume replicas should run on unique nodes
    Given mysql application is launched successfully on volume
    Then verify each volume replica gets a unique node

  Scenario: MySQL application should run when one volume replica is deleted
    Given I delete a volume replica
    Then verify mysql application is running
    And wait for "60s"
    And verify all volume replicas are running

  Scenario: MySQL application should run when other volume replica is deleted
    Given I delete another volume replica
    Then verify mysql application is running
    And wait for "60s"
    And verify all volume replicas are running
