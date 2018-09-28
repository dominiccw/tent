job "[!name!]" {
    datacenters = ["eu-west-1"]
    type = "service"

    update {
        stagger = "10s"
        max_parallel = 1
    }

    group "app" {
        count = [!group_app_size!]

        task "web" {
            driver = "docker"
            leader = true

            config {
                image = "kitematic/hello-world-nginx:latest"

                port_map {
                    http = 80
                }
            }

            resources {
                cpu = 124
                memory = 124

                network {
                    mbits = 1
                    
                    port "http" {}
                }
            }

            service {
                port = "http"

                check {
                    name = "[!job_name!] alive"
                    type = "http"
                    path = "/"
                    interval = "10s"
                    timeout = "2s"
                }
            }
        }
    }
}