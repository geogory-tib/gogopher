# This is a simple gopher protcol server written in golang!   
-------
This server is really simple but it works well for the gopher protocol I am currently running this server at gopher://geogory.space so check it out if you can!


# Config
The config for this program is super simple   
For each directory you have in your server, create a file called "gophermap"   
The gopher map is the menu for the directories there is good sources on line about using them make sure your fields are seperated by tabs and not spaces.   
In most editors there is a way to type raw control characters if your editor uses spaces instead of tabs. I will provide an example gopher map in the root of this folder   

```   
{                                                                                                                                                                                                            
  "ip":"localhost", -- the ip or hostname this is being deployed on                                                                                                                                                                              
  "port":"70", -- The port this server is being deployed on (70 is gopher's default port but any other port can work)                                                                                                                                                                                               
  "root_dir":"/var/www/gopher", -- this is the root of the sever dir where all files for your sever must be stored                                                                                                                                                            
  "log_file":"log.txt", --- this is where sever logs are stored                                                                                                                                                                                         
  "search_dir":"/search/gsearch" -- This an executable that will be ran when a user wants to query your site for a file more on this later                                                                                                                                                                               
}
```
------------
## How does the Search Feature work?   
To properly use the search feature in your gopher map create a entry with type 7 that points to the same file in your root as the search_dir   
in your config. The reason for this is so that the server can check user requests against that search directory to see if a usery wants to query a item in your server.
You don't have to use the binary I provided, you can simply make your own bash,lua or python script to do the exact thing I did. The server when It has someone make   
a request like "/search/gsearch  hello+world" will pipe in the something like this "hostname.com  70  /path/to/server/root key+words" into whatever program you have assigned in the
search_dir element in your conifg. all you have to do with your script   is print your results to stdout and the sever will capture that and display it. Unlike regular gophermaps the server does not do any formatting to the output from this command  so you will have to add the hostname and port within your script after each line of input,
make sure to also include your shebang(#!) for your interperter otherwise the server will not be able to execute your script. 

If you end up using my binary, just note that it only indexes the contents of your gophermaps through out the whole file hierarchy pulling out any non-informational text    
with any of the supplied keywords. so if a file is not in a gophermap it will not be included by the search program even if it's in your directory 


# In the works!

I am currently working on an application to make gophermaps for you so you don't have to manually config gopher maps for each dir.   
