Feature: Test high availability on Minio on Kubernetes PV
  In order to test HA on Minio on Kubernetes PV
  As an end user
  I need to be able to launch Minio on Kubernetes PV
  I need to be able to access Minio inspite of node un-availability

  Scenario: launch Minio on PV
    Given I have a kubernetes multi node cluster 
    And this cluster has volume operator installed
    When I launch minio application on volume
    Then wait for "180s"
    And verify minio application is launched successfully on volume
    And verify PVC is bound
    And verify PV is deployed

  Scenario: test high availability of minio
    Given minio application is launched successfully on volume
    Then deploy minio client config set with minio server IP
    And launch minio client put job
    And wait for "60s"
    Then verify data is put to minio server
    And cordon the node that hosts the minio pod
    And delete this minio pod
    And wait for "120s"
    Then verify minio is redeployed successfully
    And launch minio client get job
    And wait for "60s"
    And verify data is available at minio server
