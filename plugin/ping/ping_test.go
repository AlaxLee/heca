package ping

import (
	"testing"
	"github.com/spf13/viper"
)


/*
endpoint := config.GetString("endpoint")
interval := config.GetInt("jobInterval")
address := config.GetString("address")
timeout := config.GetInt("timeout")
retry   := config.GetInt("retry")
 */



func TestNewMyPing(t *testing.T) {
	v := viper.New()

	v.Set("endpoint", "lala")
	v.Set("jobInterval", 15)
	v.Set("address", "1.2.3.4")
	v.Set("timeout", 5)
	v.Set("retry", 3)

	m, err := NewMyPing(v)
	if err != nil {
		t.Fatal("create MyPing object failed", err)
	} else {
		if m.endpoint == v.GetString("endpoint") &&
			m.interval == v.GetInt("jobInterval") &&
			m.address == v.GetString("address") &&
			m.timeout == v.GetInt("timeout") &&
			m.retry == v.GetInt("retry") {
			t.Log("OK")
		} else {
			t.Fatal("create MyPing object failed")
		}
	}
}



func TestMyPing_Ping(t *testing.T) {
	v := viper.New()
	v.Set("endpoint", "lala")
	v.Set("jobInterval", 15)
	v.Set("address", "127.0.0.1")
	v.Set("timeout", 5)
	v.Set("retry", 3)
	m, err := NewMyPing(v)
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Ping()
	if err != nil {
		t.Fatal(err)
	}
}


func TestMyPing_Do(t *testing.T) {
	v := viper.New()
	v.Set("endpoint", "lala")
	v.Set("jobInterval", 15)
	v.Set("address", "127.0.0.1")
	v.Set("timeout", 5)
	v.Set("retry", 3)
	m, err := NewMyPing(v)
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan interface{}, 10)
	resultChan := (chan<- interface{})(c)
	m.Do(resultChan)



	/*
	resultChan <- map[string]interface{} {
		"endpoint": p.endpoint,
		"metric": "ping.available",
		"timestamp": time.Now().Unix(),
		"step": p.interval,
		"value": p.result.available,
		"counterType": "GAUGE",
		"tags": "",
	}
	 */
	available, ok := <- c
	if !ok {
		t.Fatal("get available failed")
	}
	if a, ok := available.(map[string]interface{}); !ok {
		t.Fatal("available is not a map[string]interface{}")
	} else {

		if endpoint, ok := a["endpoint"]; !ok {
			t.Error("available must have endpoint")
		} else if endpoint != v.GetString("endpoint") {
			t.Error("available's endpoint has changed")
		}

		if metric, ok := a["metric"]; !ok {
			t.Error("available must have metric")
		} else if metric != "ping.available" {
			t.Error("available's metric has changed")
		}

		if _, ok := a["timestamp"]; !ok {
			t.Error("available must have timestamp")
		}

		if step, ok := a["step"]; !ok {
			t.Error("available must have step")
		} else if step != v.GetInt("jobInterval") {
			t.Error("available's step has changed")
		}

		if _, ok := a["value"]; !ok {
			t.Error("available must have value")
		}

		if counterType, ok := a["counterType"]; !ok {
			t.Error("available must have counterType")
		} else if counterType != "GAUGE" {
			t.Error("available's counterType has changed")
		}

		if tags, ok := a["tags"]; !ok {
			t.Error("available must have tags")
		} else if tags != "" {
			t.Error("available's tags has changed")
		}
	}

	/*
	resultChan <- map[string]interface{} {
		"endpoint": p.endpoint,
		"metric": "ping.delay",
		"timestamp": time.Now().Unix(),
		"step": p.interval,
		"value": p.result.delay,
		"counterType": "GAUGE",
		"tags": "unit=second",
	}
	 */
	delay, ok := <- c
	if !ok {
		t.Fatal("get delay failed")
	}
	if d, ok := delay.(map[string]interface{}); !ok {
		t.Fatal("delay is not a map[string]interface{}")
	} else {

		if endpoint, ok := d["endpoint"]; !ok {
			t.Error("delay must have endpoint")
		} else if endpoint != v.GetString("endpoint") {
			t.Error("delay's endpoint has changed")
		}

		if metric, ok := d["metric"]; !ok {
			t.Error("delay must have metric")
		} else if metric != "ping.delay" {
			t.Error("delay's metric has changed")
		}

		if _, ok := d["timestamp"]; !ok {
			t.Error("delay must have timestamp")
		}

		if step, ok := d["step"]; !ok {
			t.Error("delay must have step")
		} else if step != v.GetInt("jobInterval") {
			t.Error("delay's step has changed")
		}

		if _, ok := d["value"]; !ok {
			t.Error("delay must have value")
		}

		if counterType, ok := d["counterType"]; !ok {
			t.Error("delay must have counterType")
		} else if counterType != "GAUGE" {
			t.Error("delay's counterType has changed")
		}

		if tags, ok := d["tags"]; !ok {
			t.Error("delay must have tags")
		} else if tags != "unit=second" {
			t.Error("delay's tags has changed")
		}
	}

	/*
	resultChan <- map[string]interface{} {
		"endpoint": p.endpoint,
		"metric": "ping.loss",
		"timestamp": time.Now().Unix(),
		"step": p.interval,
		"value": p.result.loss,
		"counterType": "GAUGE",
		"tags": "unit=percent",
	}
	 */
	loss, ok := <-c
	if !ok {
		t.Fatal("get loss failed")
	}
	if l, ok := loss.(map[string]interface{}); !ok {
		t.Fatal("loss is not a map[string]interface{}")
	} else {

		if endpoint, ok := l["endpoint"]; !ok {
			t.Error("loss must have endpoint")
		} else if endpoint != v.GetString("endpoint") {
			t.Error("loss's endpoint has changed")
		}

		if metric, ok := l["metric"]; !ok {
			t.Error("loss must have metric")
		} else if metric != "ping.loss" {
			t.Error("loss's metric has changed")
		}

		if _, ok := l["timestamp"]; !ok {
			t.Error("loss must have timestamp")
		}

		if step, ok := l["step"]; !ok {
			t.Error("loss must have step")
		} else if step != v.GetInt("jobInterval") {
			t.Error("loss's step has changed")
		}

		if _, ok := l["value"]; !ok {
			t.Error("loss must have value")
		}

		if counterType, ok := l["counterType"]; !ok {
			t.Error("loss must have counterType")
		} else if counterType != "GAUGE" {
			t.Error("loss's counterType has changed")
		}

		if tags, ok := l["tags"]; !ok {
			t.Error("loss must have tags")
		} else if tags != "unit=percent" {
			t.Error("loss's tags has changed")
		}
	}

}
