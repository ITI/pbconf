/*
 * Ontology configuration gives a way of specifying some required core ontology
 * properties, and extending to allow any secondary ontologies.  json files must follow
 * the format :

{
    "prefixes":[
        {"prefix":"owl","iri":"http://www.w3.org/2002/07/owl"},
        {"prefix":"pbconf","iri":"http://iti.illinois.edu/iti/pbconf/core"},
        {"prefix":"config","iri":"http://iti.illinois.edu/iti/pbconf/config"},
        {"prefix":"policy","iri":"http://iti.illinois.edu/iti/pbconf/policy"}       
    ],
    "ontologyDirectory":"owl",
    "coreOntology":"pbconf",
    "closedWorldOntology":"policy",
    "additionalOntologies":[
        "config",
        "http://iti.illinois.edu/iti/pbconf/policy",
        "file:owl/pbtest.owl",
        "file:owl/pbtest2.ttl"
    ],
    "closedWorldReasoners": {
        "IndividualMustExist":"edu.illinois.iti.pbconf.ontology.validator.IndividualMustExist",
        "ClassMustHaveProperty":"edu.illinois.iti.pbconf.ontology.validator.ClassMustHaveProperty",
        "ClassMustNotHaveProperty":"edu.illinois.iti.pbconf.ontology.validator.ClassMustNotHaveProperty"
    }
}

Where ontologies can be any length including 0, of prefix / iri pairs

 */
package edu.illinois.iti.pbconf.ontology;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;
import org.apache.log4j.Logger;
import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

/**
 * Ontology configuration class
 * Load constant values from JSON files, recover in situations where that fails
 * @author Joe
 */
abstract public class OntologyConfig {
    Logger logger = Logger.getLogger(OntologyConfig.class.getName().replaceFirst(".+\\.",""));
      
    /**
     * Get a configuration property by name
     * @param property
     * @return
     */
    abstract public String get(String property);

    /**
     * Set a configuration property by name
     * @param property
     * @param value
     */
    abstract public void set(String property, String value);

    /**
     * remove a configuration property
     * @param property
     * @return
     */
    abstract public boolean remove(String property);

    /**
     * Get the ontology prefixes known to the system
     * @return
     */
    abstract public Map<String, String> getPrefixes();

    /**
     * Add an ontology prefix to the system
     * @param prefix
     * @param iri
     */
    abstract public void addPrefix(String prefix, String iri);

    /**
     * Get ontology beyond the core ontology
     * @return
     */
    abstract public ArrayList getAdditionalOntologies();

    /**
     * Get prefixes of ontology, optional including root ontology
     * @param includeRoot
     * @return
     */
    abstract public ArrayList getAdditionalOntologyPrefixes(boolean includeRoot);

    /**
     * Get closed world reasoner modules
     * @return
     */
    abstract public Map<String, String> getClosedWorldReasoners();

    /**
     * Get the closed world reasoner prefix string
     * @return
     */
    abstract public String getClosedWorldPrefixStr();

    /**
     * Get the policy ontology string
     * @return
     */
    abstract public String getPolicyPrefix();

    /**
     * Get the configuration ontology prefix string
     * @return
     */
    abstract public String getConfigPrefix();

    /**
     * Get the partial configuration ontology prefix string
     * @return
     */
    abstract public String getPartialPrefix();
    
    /**
     * Get the core ontology prefix string
     * @return
     */
    abstract public String getCorePrefix();
     
    /**
     * Takes a configuration file name and returns the full file path
     * @param configFileName
     * @return configFilePath
     */
    public static String getConfigPath(String configFileName) {
        //String filePath = System.getProperty("user.dir");
        //filePath = filePath.concat("/config/");
        //filePath = filePath.concat(configFileName);
        String filePath = configFileName;
        return filePath;
    }
    
    /**
     * Load original default configuration into a given JSONObject
     * @param obj
     */
    public static void loadDefaultConfig(JSONObject obj) {
        obj.put("ontologyDirectory", "");
        obj.put("coreOntology", "");
        obj.put("closedWorldOntology", "");
        obj.put("prefixes", new JSONArray());
        obj.put("additionalOntologies", new JSONArray());
    }

    /**
     * JSON Configuration class
     * Used for loading configuration by way of JSON file
     */
    public static class JSONConfig extends OntologyConfig {
        static JSONObject configObj = new JSONObject();
        
        /**
         * Get a mapped string property by key
         * @param property
         * @return 
         */
        @Override
        public String get(String property) {
            if (configObj.has(property)) {
                return configObj.getString(property);
            }
            return null;
        }
        
        /**
         * Set a String,String property for later access
         * @param property
         * @param value 
         */
        @Override
        public void set(String property, String value) {
            configObj.put(property, value);
        }
        
        /**
         * Remove a non-essential property
         * Essential properties will be skipped
         * @param property 
         * @return  
         */
        @Override
        public boolean remove(String property) {
            if ("coreOntology".equals(property) || "closedWorldOntology".equals(property) || "ontologyDirectory".equals(property) || "prefixes".equals(property) || "additionalOntologies".equals(property)) {
                return false;
            }
            
            if (configObj.has(property)) {
                configObj.remove(property);
                return true;
            }
            
            return false;
        }
        
        /**
         * Return a HashMap of prefix / IRI combos
         * @return 
         */
        @Override
        public Map<String, String> getPrefixes() {
            if (configObj.has("prefixes")) {
                JSONArray jArr = configObj.getJSONArray("prefixes");
                Map<String, String> map = new HashMap<>();                           
                for (int i = 0; i < jArr.length(); i++) {
                    map.put(jArr.getJSONObject(i).getString("prefix"), jArr.getJSONObject(i).getString("iri"));
                }      
                return map;
            }
            return null;
        }
        
        /**
         * Add a prefix to an existing configuration at run time
         * @param prefix
         * @param iri 
         */
        @Override
        public void addPrefix(String prefix, String iri) {
            boolean found = false;
            if (configObj.has("prefixes")) {
                JSONArray prefixes = configObj.getJSONArray("prefixes");
                for (int i = 0; i < prefixes.length(); i++) {
                    //If this matches, we're actually updating, not adding
                    if (prefixes.getJSONObject(i).getString("prefix").equals(prefix)) {
                        found = true;
                        prefixes.getJSONObject(i).put(prefix, iri);
                    }
                }
                if (found == false) {
                    JSONObject newItem = new JSONObject();
                    newItem.put(prefix, iri);
                    prefixes.put(newItem);
                }
            }
        }
        
        /**
         * Get the JSONArray additional Ontologies
         * @return 
         */
        @Override
        public ArrayList getAdditionalOntologies() {
            JSONArray jArr = configObj.getJSONArray("additionalOntologies");
            ArrayList arr = new ArrayList();
            for (int i = 0; i < jArr.length(); i++) {
                arr.add(jArr.getString(i));
            }
            return arr;
        }
        
        /**
         * Get an array list containing all the prefixes of loaded ontologies
         * @param includeRoot
         * @return 
         */
        @Override
        public ArrayList getAdditionalOntologyPrefixes(boolean includeRoot) {
            ArrayList arr = new ArrayList();
            JSONArray addOnts = configObj.getJSONArray("additionalOntologies");
            JSONArray jPrefixes = configObj.getJSONArray("prefixes");

            for (int i = 0; i < addOnts.length(); i++) {
                String addOnt = addOnts.getString(i);
                String resultingPrefix = "";
                boolean found = false;
                for (int j = 0; j < jPrefixes.length(); j++) {
                    if (found == true) {
                        continue;
                    }
                    //Each prefix in the array is a JSON object containing iri, prefix
                    JSONObject prefix = jPrefixes.getJSONObject(j);
                    if (addOnt.equals(prefix.getString("prefix")) || addOnt.equals(prefix.getString("iri"))) {
                        resultingPrefix = prefix.getString("prefix");
                        found = true;
                    }
                }
                if (found == true && !"".equals(resultingPrefix)) {
                    arr.add(resultingPrefix);
                }
            }
            
            if (Ontology.instance().getIsConfigurationLoaded()) {
                if (!Arrays.asList(arr).contains(Ontology.instance().getConfig().getConfigPrefix())) {
                    arr.add(Ontology.instance().getConfig().getConfigPrefix());
                }
            }
             
            if (Ontology.instance().getIsPartialConfigurationLoaded()) {
                if (!Arrays.asList(arr).contains(Ontology.instance().getConfig().getPartialPrefix())) {
                    arr.add(Ontology.instance().getConfig().getPartialPrefix());
                }
            }
            
            if (includeRoot == true) {
                arr.add(configObj.getString("coreOntology"));
            }
            
            return arr;
        }
        
        /**
         * Get a Map of closed world reasoner(s), where Map key is reasoner "name", 
         * and value is reasoner 
         * @return 
         */
        @Override
        public Map<String,String> getClosedWorldReasoners() {
            JSONObject jsonReasoners = configObj.getJSONObject("closedWorldReasoners");
            HashMap<String,String> reasoners = new HashMap<>();
            if (jsonReasoners != null) {
                Set<String> keys = jsonReasoners.keySet();
                for (String key : keys) {
                    String className = jsonReasoners.getString(key);
                    if (className != null) {
                        reasoners.put(key, className);
                    }
                }
            }
            return reasoners;
        }
        
        /**
         * Get the ontology iri associated with the ClosedWorldReasoning
         * @return 
         */
        @Override
        public String getClosedWorldPrefixStr() {
            String ontPrefix = configObj.getString("closedWorldOntology");
            for (Map.Entry<String, String> prefix : getPrefixes().entrySet()) {
                if (prefix.getKey().equals(ontPrefix)) {
                    return prefix.getValue();
                }
            }
            return "";
        }
        
        /**
         * Get the prefix for the policy ontology from the configuration file
         * @return 
         */
        @Override
        public String getPolicyPrefix() {
            return configObj.getString("policyOntology");
        }
        
        /**
         * Get the configuration ontology prefix from the configuration file
         * @return 
         */
        @Override
        public String getConfigPrefix() {
            return configObj.getString("configOntology");
        }

        /**
         * Get the partial configuration ontology prefix from the configuration file
         * @return 
         */
        @Override
        public String getPartialPrefix() {
            return configObj.getString("partialConfigOntology");
        }
        
        /**
         * Get the core ontology prefix from the configuration file
         * @return 
         */
        @Override
        public String getCorePrefix() {
            return configObj.getString("coreOntology");
        }
        
        /**
         * 
         * @param configFile 
         */
        private void initialize(String configFile) {
            try {
                String filePath;
                filePath = getConfigPath(configFile);
                File jsonFile = new File(filePath);
                byte[] encoded = Files.readAllBytes(jsonFile.toPath());
                String json = new String(encoded, "UTF-8");
                configObj = new JSONObject(json);
            } catch (JSONException | IOException e) {
                logger.error(e);
                loadDefaultConfig(configObj);
            }
        }
        
        /**
         * With an empty constructor, load the base configuration only
         */
        public JSONConfig() {
            String c = System.getProperty("user.dir");
            c  = c.concat("/config/pbconf_base.json");
            initialize(c);
        }
        
        /**
         * Constructor using a JSON formatted file
         * @param configFile 
         */
        public JSONConfig(String configFile){ 
            initialize(configFile);
        }
    }
}
