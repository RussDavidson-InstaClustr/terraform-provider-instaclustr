provider "instaclustr" {
  username     = "%s"
  api_key      = "%s"
  api_hostname = "%s"
}


resource "instaclustr_cluster" "validOpenSearch" {
  cluster_name            = "tf-opensearch-test"
  data_centre             = "US_WEST_2"
  sla_tier                = "NON_PRODUCTION"
  cluster_network         = "192.168.0.0/18"
  private_network_cluster = false
  cluster_provider        = {
    name = "AWS_VPC",
  }
  rack_allocation         = {
    number_of_racks = 3
    nodes_per_rack  = 2
  }
  bundle {
    bundle  = "OPENSEARCH"
    version = "1.3.7"
    options = {
      dedicated_master_nodes  = false
      master_node_size        = "SRH-DEV-t4g.small-5"
      data_node_size          = "SRH-DEV-t4g.small-5"
      security_plugin         = true
      index_management_plugin = true
      knn_plugin              = true
      reporting_plugin        = true
      icu_plugin              = true
      alerting_plugin         = false
      notifications_plugin    = false
      client_encryption       = true
    }
  }
}
