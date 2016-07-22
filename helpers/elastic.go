package helpers

import (
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"log"
	"time"
)

func Check_index_and_create(index string) bool {
	mapping := `{"settings":{"number_of_shards":1,"number_of_replicas":0}}`
	client, err := elastic.NewClient()
	if err != nil {
		log.Println("error connecting to ES", err)
		return false
	}
	fmt.Printf("Checking for index: %v now \n", index)
	exists, err := client.IndexExists(index).Do()
	if err != nil {
		log.Println("error checking if index exists", err)
		return false
	}
	if !exists {
		_, err = client.CreateIndex(index).BodyString(mapping).Do()
		if err != nil {
			log.Println("erorr creating index", err)
			return false
		}
		fmt.Println("Sleeping to allow ES to allocate indexes")
		time.Sleep(5 * time.Second)
	}

	return true
}

func checkCreateSonaHostsSSLIndex() {
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	//let's check if index exists:
	exists, err := client.IndexExists("passive-ssl-sonar-hosts").Do()
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		mapping := `{
    "settings":{
        "number_of_shards":5,
        "number_of_replicas":0
    },
    "mappings":{
         "host" : {
        "properties" : {
          "host": {"type": "ip", "index": "analyzed"},
          "hash": {"type": "string"},
          "first_seen": {"type": "date", "format": "dateOptionalTime"},
          "last_seen": {"type": "date", "format": "dateOptionalTime"},
          "asn": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "country_code": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "city": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "region": {"type": "string", "analyzer": "keyword", "index": "analyzed"},
          "port": {"type": "integer"},
          "source": {"type": "string"}
        }
      }
        }
    }
}`
		_, err = client.CreateIndex("passive-ssl-sonar-hosts").BodyString(mapping).Do()
		if err != nil {
			panic(err)
		}
		fmt.Println("Index Created")
		fmt.Println("Sleeping to allow ES to allocate indexes")
		time.Sleep(5 * time.Second)
		return
	}
	fmt.Println("The index already existed")
	return
}



func checkCreateSonarCertsSSLIndex() {
	client, err := elastic.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	//let's check if index exists:
	exists, err := client.IndexExists("passive-ssl-sonar-certs").Do()
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		mapping := `{
    "settings":{
        "number_of_shards":5,
        "number_of_replicas":0
    },
        "mappings": {
      "certificate": {
        "properties": {
          "extensions": {
            "properties": {
              "authority_info_access": {
                "properties": {
                  "issuer_urls": {
                    "type": "string"
                  },
                  "ocsp_urls": {
                    "type": "string"
                  }
                }
              },
              "authority_key_id": {
                "type": "string"
              },
              "basic_constraints": {
                "properties": {
                  "is_ca": {
                    "type": "boolean"
                  },
                  "max_path_len": {
                    "type": "long"
                  }
                }
              },
              "certificate_policies": {
                "type": "string"
              },
              "crl_distribution_points": {
                "type": "string"
              },
              "extended_key_usage": {
                "type": "long"
              },
              "key_usage": {
                "properties": {
                  "certificate_sign": {
                    "type": "boolean"
                  },
                  "content_commitment": {
                    "type": "boolean"
                  },
                  "crl_sign": {
                    "type": "boolean"
                  },
                  "data_encipherment": {
                    "type": "boolean"
                  },
                  "decipher_only": {
                    "type": "boolean"
                  },
                  "digital_signature": {
                    "type": "boolean"
                  },
                  "encipher_only": {
                    "type": "boolean"
                  },
                  "key_agreement": {
                    "type": "boolean"
                  },
                  "key_encipherment": {
                    "type": "boolean"
                  },
                  "value": {
                    "type": "long"
                  }
                }
              },
              "name_constraints": {
                "properties": {
                  "critical": {
                    "type": "boolean"
                  },
                  "permitted_names": {
                    "type": "string",
                    "fields": {
                      "raw": {
                        "type": "string",
                        "index": "not_analyzed"
                      }
                    }
                  }
                }
              },
              "subject_alt_name": {
                "properties": {
                  "dns_names": {
                    "type": "string",
                    "fields": {
                      "raw": {
                        "type": "string",
                        "index": "not_analyzed"
                      }
                    }
                  },
                  "email_addresses": {
                    "type": "string",
                    "fields": {
                      "raw": {
                        "type": "string",
                        "index": "not_analyzed"
                      }
                    }
                  },
                  "ip_addresses": {
                    "type": "string",
                    "fields": {
                      "raw": {
                        "type": "string",
                        "index": "not_analyzed"
                      }
                    }
                  }
                }
              },
              "subject_key_id": {
                "type": "string"
              }
            }
          },
          "fingerprint_md5": {
            "type": "string"
          },
          "fingerprint_sha1": {
            "type": "string"
          },
          "fingerprint_sha256": {
            "type": "string"
          },
          "issuer": {
            "properties": {
              "common_name": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "country": {
                "type": "string"
              },
              "locality": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "organization": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "organizational_unit": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "postal_code": {
                "type": "string"
              },
              "province": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "serial_number": {
                "type": "string"
              },
              "street_address": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              }
            }
          },
          "issuer_dn": {
            "type": "string",
            "fields": {
              "raw": {
                "type": "string",
                "index": "not_analyzed"
              }
            }
          },
          "serial_number": {
            "type": "string"
          },
          "signature": {
            "properties": {
              "self_signed": {
                "type": "boolean"
              },
              "signature_algorithm": {
                "properties": {
                  "name": {
                    "type": "string"
                  },
                  "oid": {
                    "type": "string"
                  }
                }
              },
              "valid": {
                "type": "boolean"
              },
              "value": {
                "type": "string"
              }
            }
          },
          "signature_algorithm": {
            "properties": {
              "name": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "oid": {
                "type": "string"
              }
            }
          },
          "subject": {
            "properties": {
              "common_name": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "country": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "locality": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "organization": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "organizational_unit": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "postal_code": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "province": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              },
              "serial_number": {
                "type": "string"
              },
              "street_address": {
                "type": "string",
                "fields": {
                  "raw": {
                    "type": "string",
                    "index": "not_analyzed"
                  }
                }
              }
            }
          },
          "subject_dn": {
            "type": "string",
            "fields": {
              "raw": {
                "type": "string",
                "index": "not_analyzed"
              }
            }
          },
          "subject_key_info": {
            "properties": {
              "dsa_public_key": {
                "properties": {
                  "g": {
                    "type": "string"
                  },
                  "p": {
                    "type": "string"
                  },
                  "q": {
                    "type": "string"
                  },
                  "y": {
                    "type": "string"
                  }
                }
              },
              "ecdsa_public_key": {
                "properties": {
                  "b": {
                    "type": "string"
                  },
                  "gx": {
                    "type": "string"
                  },
                  "gy": {
                    "type": "string"
                  },
                  "n": {
                    "type": "string"
                  },
                  "p": {
                    "type": "string"
                  },
                  "x": {
                    "type": "string"
                  },
                  "y": {
                    "type": "string"
                  }
                }
              },
              "key_algorithm": {
                "properties": {
                  "name": {
                    "type": "string"
                  },
                  "oid": {
                    "type": "string"
                  }
                }
              },
              "rsa_public_key": {
                "properties": {
                  "exponent": {
                    "type": "long"
                  },
                  "length": {
                    "type": "long"
                  },
                  "modulus": {
                    "type": "string"
                  }
                }
              }
            }
          },
          "unknown_extensions": {
            "properties": {
              "critical": {
                "type": "boolean"
              },
              "id": {
                "type": "string"
              },
              "value": {
                "type": "string"
              }
            }
          },
          "validity": {
            "properties": {
              "end": {
                "type": "date",
                "format": "strict_date_optional_time||epoch_millis"
              },
              "start": {
                "type": "date",
                "format": "strict_date_optional_time||epoch_millis"
              }
            }
          },
          "version": {
            "type": "long"
          }
        }
      }
    }
    }`
		_, err = client.CreateIndex("passive-ssl-sonar-certs").BodyString(mapping).Do()
		if err != nil {
			panic(err)
		}
		fmt.Println("Index Created")
		fmt.Println("Sleeping to allow ES to allocate indexes")
		time.Sleep(5 * time.Second)
		return
	}
	fmt.Println("The index already existed")
	return
}