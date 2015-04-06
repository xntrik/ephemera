package main

import (
	"code.google.com/p/goauth2/oauth"
	"flag"
	"github.com/digitalocean/godo"
	"io/ioutil"
	"log"
	"os/user"
)

type Config struct {
	key  string
	name string
}

// These are sane defaults for me right now, but they should be generalised or
// pulled into a config file or something
const (
	REGION      = "sfo1"
	SIZE        = "2gb"
	IMAGE_SLUG  = "debian-7-0-x64"
	FINGERPRINT = "91:ff:af:1c:e2:0c:5e:b7:dd:8d:6c:27:0d:e6:20:63"
)

func parse_flags() *Config {
	name := flag.String("name", "", "instance name")
	flag.Parse()

	if *name == "" {
		return nil
	}

	return &Config{
		"",
		*name,
	}
}

func get_token() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	dat, err := ioutil.ReadFile(dir + "/.ephemera")
	if err != nil {
		log.Fatal("Couldn't read token file")
	}
	return string(dat)
}

func get_config() *Config {
	base := parse_flags()
	if base == nil {
		return nil
	}

	base.key = get_token()

	return base
}

func main() {
	cfg := get_config()
	if cfg == nil {
		log.Fatal("Couldn't parse flags")
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: cfg.key},
	}
	client := godo.NewClient(t.Client())
	// Specialcase to dump all instance types. This is silly
	if cfg.name == "?" {
		list_all_images(client)
		return
	}

	droplet := create_ephemeral_instance(client, cfg.name)
	log.Printf("%s", droplet)
	log.Println("Waiting for droplet to come up")
}

func create_ephemeral_instance(client *godo.Client, name string) *godo.DropletRoot {

	createRequest := &godo.DropletCreateRequest{
		Name:    name,
		Region:  REGION,
		Size:    SIZE,
		SSHKeys: []godo.DropletCreateSSHKey{{Fingerprint: FINGERPRINT}},
		Image: godo.DropletCreateImage{
			Slug: IMAGE_SLUG,
		},
	}

	newDroplet, _, err := client.Droplets.Create(createRequest)

	if err != nil {
		log.Fatal("Something bad happened: %s", err)
	}

	return newDroplet
}

func list_all_images(client *godo.Client) {
	opt := &godo.ListOptions{}
	for {
		images, resp, err := client.Images.List(opt)
		if err != nil {
			log.Fatal(err)
		}

		// append the current page's droplets to our list
		for _, d := range images {
			log.Printf("- %s\n", d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			log.Fatal(err)
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

}
