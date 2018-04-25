/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import java.util.ArrayList;
import java.util.List;
import org.apache.log4j.Logger;

/**
 * Used to generate example configuration and policy data requests for use with unit tests.
 * @author josephdigiovanna
 */
public class DataGenerator {

    static Logger logger = Logger.getLogger(DataGenerator.class.getName().replaceFirst(".+\\.", ""));

    public static String getPasswordPolicy() {
        List dataArray = new ArrayList();
        String[] axioms = new String[3];
        axioms[0] = "{\"s\":\"password.level1\",\"p\":\"min-length\",\"o\":\"4\"}";
        axioms[1] = "{\"s\":\"password.level1\",\"p\":\"max-length\",\"o\":\"16\"}";
        axioms[2] = "{\"s\":\"password.level1\",\"p\":\"complexity\",\"o\":\"LOWERCASE\"}";
        
        dataArray.add(DataGenerator.buildJSONPolicyRequestStr("SEL421", axioms));
        return DataGenerator.buildFullJSONPolicyRequestStr("SEL421", dataArray);
    }
    
    /**
     * This creates both SEL421 and LINUX device data in the configuration ontology
     * This is used only for unit tests.
     * @return
     */
    public static List getTestConfigurationDataRequestBoth() {
        /**
         * Setup 2 complete devices in the configuration ontology for testing
         * we'll try to use some names that won't interfere with anything else
         * This is so that we don't have to keep any actual devices in the
         * ontologies that could potentially interfere with policy we set for
         * other stuff and cause inconsistencies
         */
        List deviceArr = new ArrayList();
        List indArr = new ArrayList();

        //new IP address
        deviceArr.add(buildJSONRequestStr("type", "", "IPAddress", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte1", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte2", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte3", "1", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte4", "10", ""));
       
        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICE_test_port5_ipaddrA", deviceArr));
        
        //SELPort5 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "SEL421Port5", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAllowedOperation", "sel421_automation_setting", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasDNP3Stt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasFTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasIPAddr", "sel421FAKEDEVICE_test_port5_ipaddrA", ""));

        //Data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasMacAddr", "aa:bb:cc:dd:ee:ff", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasRetryDelayed", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "5", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetPort", "23", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetTimeout", "10", ""));
        
        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEA_port5", deviceArr));

        //SEL421 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("service", "hasAccessLoggingStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAlarmStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasGPSStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));

        //SEL421 data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "20", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1APwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "dsodinbqeor", ""));

        //Set hasPort5 and Type
        deviceArr.add(buildJSONRequestStr("service", "hasPort5", "sel421FAKEDEVICEA_port5", ""));
        deviceArr.add(buildJSONRequestStr("type", "", "SEL421", ""));

        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEA", deviceArr));
        
        //SELPort5 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("variable", "type", "SEL421Port5", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAllowedOperation", "sel421_automation_setting", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasDNP3Stt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasFTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasIPAddr", "sel421FAKEDEVICE_test_port5_ipaddrA", ""));

        //Data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasMacAddr", "aa:bb:cc:dd:ee:ff", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasRetryDelayed", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "5", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetPort", "23", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetTimeout", "10", ""));

        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB_port5", deviceArr));
        
        //SEL421 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("service", "hasAccessLoggingStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAlarmStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasGPSStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));

        //SEL421 data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "20", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1APwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "dsodinbqeor", ""));

        //Set hasPort5 and Type
        deviceArr.add(buildJSONRequestStr("service", "hasPort5", "sel421FAKEDEVICEB_port5", ""));
        deviceArr.add(buildJSONRequestStr("type", "", "SEL421", ""));
        
        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", deviceArr));
        
        //Create people
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Joe", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Adrian", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Andy", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Jeremy", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Ashwini", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "LINUX", ""));       
        deviceArr.add(buildJSONRequestStr("service", "authusers", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "anonymousftp", "on", ""));
        deviceArr.add(buildJSONRequestStr("service_option", "authusers", "Joe Andy Adrian Jeremy Ashwini", "telnet"));
        deviceArr.add(buildJSONRequestStr("service_option", "authusers", "Joe Andy Adrian Jeremy Ashwini", "ftp"));
        deviceArr.add(buildJSONRequestStr("service_option", "authusers", "Adrian Jeremy Ashwini", "dnp3"));
        deviceArr.add(buildJSONRequestStr("service_option", "anonymousftp", "on", ""));       
        indArr.add(buildFullJSONRequestStr("LINUX", "linuxa", deviceArr));
        
        return indArr;
    }
    
    /**
     * Create only the LINUX specific ontology data
     * @return
     */
    public static List getTestConfigurationDataRequestLINUX() {
        List deviceArr = new ArrayList();
        List indArr = new ArrayList();
        
        //Create people
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Joe", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Adrian", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Andy", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Jeremy", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "Person", ""));
        indArr.add(buildFullJSONRequestStr("LINUX", "Ashwini", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "LINUX", ""));       
        deviceArr.add(buildJSONRequestStr("service", "authusers", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "anonymousftp", "on", ""));
        deviceArr.add(buildJSONRequestStr("service_option", "authusers", "Joe Andy Adrian Jeremy Ashwini", "telnet"));
        deviceArr.add(buildJSONRequestStr("service_option", "authusers", "Joe Andy Adrian Jeremy Ashwini", "ftp"));
        deviceArr.add(buildJSONRequestStr("service_option", "authusers", "Adrian Jeremy Ashwini", "dnp3"));
        deviceArr.add(buildJSONRequestStr("service_option", "anonymousftp", "on", ""));       
        indArr.add(buildFullJSONRequestStr("LINUX", "linuxa", deviceArr));
        
        return indArr;
    }
    
    /**
     * Create only the SEL421 device data
     * @return
     */
    public static List getTestConfigurationDataRequestSEL421() {
        List deviceArr = new ArrayList();
        List indArr = new ArrayList();

        //new IP address
        deviceArr.add(buildJSONRequestStr("type", "", "IPAddress", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte1", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte2", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte3", "1", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasByte4", "10", ""));
       
        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICE_test_port5_ipaddrA", deviceArr));
        
        //SELPort5 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "SEL421Port5", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAllowedOperation", "sel421_automation_setting", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasDNP3Stt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasFTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasIPAddr", "sel421FAKEDEVICE_test_port5_ipaddrA", ""));

        //Data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasMacAddr", "aa:bb:cc:dd:ee:ff", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasRetryDelayed", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "5", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetPort", "23", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetTimeout", "10", ""));
        
        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEA_port5", deviceArr));

        //SEL421 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("service", "hasAccessLoggingStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAlarmStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasGPSStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));

        //SEL421 data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "20", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1APwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "dsodinbqeor", ""));

        //Set hasPort5 and Type
        deviceArr.add(buildJSONRequestStr("service", "hasPort5", "sel421FAKEDEVICEA_port5", ""));
        deviceArr.add(buildJSONRequestStr("type", "", "SEL421", ""));

        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEA", deviceArr));
        
        //SELPort5 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("variable", "type", "SEL421Port5", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAllowedOperation", "sel421_automation_setting", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasDNP3Stt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasFTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasIPAddr", "sel421FAKEDEVICE_test_port5_ipaddrA", ""));

        //Data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasMacAddr", "aa:bb:cc:dd:ee:ff", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasRetryDelayed", "10", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "5", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetPort", "23", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasTelnetTimeout", "10", ""));

        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB_port5", deviceArr));
        
        //SEL421 object properties
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("service", "hasAccessLoggingStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasAlarmStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasGPSStt", "on", ""));
        deviceArr.add(buildJSONRequestStr("service", "hasNTPStt", "on", ""));

        //SEL421 data properties
        deviceArr.add(buildJSONRequestStr("variable", "hasAccessTimeout", "20", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1APwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "dsodinbqeor", ""));

        //Set hasPort5 and Type
        deviceArr.add(buildJSONRequestStr("service", "hasPort5", "sel421FAKEDEVICEB_port5", ""));
        deviceArr.add(buildJSONRequestStr("type", "", "SEL421", ""));
        
        indArr.add(buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", deviceArr));

        return indArr;
    }
    
    /**
     * Create CWR data
     * @return
     */
    public static List getTestConfigurationDataRequestCWR() {
        List indArr = new ArrayList();
        
        List deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "TEST421", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "testPass", ""));       
        indArr.add(buildFullJSONRequestStr("SEL421", "test421a", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "TEST421", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "testPass", ""));      
        indArr.add(buildFullJSONRequestStr("SEL421", "test421b", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "TEST421", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1Pwd", "testPassAlt", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPassAlt", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "testPassAlt", ""));       
        indArr.add(buildFullJSONRequestStr("SEL421", "test421a", deviceArr));
        
        deviceArr = new ArrayList();
        deviceArr.add(buildJSONRequestStr("type", "", "TEST421", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl1Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvl2Pwd", "testPass", ""));
        deviceArr.add(buildJSONRequestStr("variable", "hasLvlCPwd", "testPass", ""));
        indArr.add(buildFullJSONRequestStr("SEL421", "test421c", deviceArr));
        
        return indArr;
    }
    
    /**
     * Wrap multiple individuals into one configuration
     * 
     * @param dataArray
     * @return 
     */
    public static String buildOuterJSONRequestStr(List dataArray) {
        String jsonStr = "{\"ontology\":\"config\", \"properties\":[";
        
        for (int i = 0; i < dataArray.size(); i++) {
            jsonStr = jsonStr.concat((String) dataArray.get(i));
            if (i != dataArray.size() - 1) {
                jsonStr = jsonStr.concat(",");
            }
        }

        jsonStr += "]}";

        return jsonStr;
    }

    /**
     * Build out a full configuration statement so we have real configuration data
     * @param ontologizer
     * @param ind
     * @param dataArray
     * @return 
     */
    public static String buildFullJSONRequestStr(String ontologizer, String ind, List dataArray) {
        String jsonStr = "{\"ontology\":\"config\", \"ontologizer\":\"" + ontologizer + "\",\"individual\":\"" + ind + "\",\"properties\":[";
        
        for (int i = 0; i < dataArray.size(); i++) {
            jsonStr = jsonStr.concat((String) dataArray.get(i));
            if (i != dataArray.size() - 1) {
                jsonStr = jsonStr.concat(",");
            }
        }
        
        jsonStr += "]}";
        
        return jsonStr;
    }
    
    /**
     * Construct a JSON request for parsing, for a configuration operation
     * @param Op
     * @param Key
     * @param Val
     * @param Svc
     * @return 
     */
    public static String buildJSONRequestStr(String Op, String Key, String Val, String Svc) {
        String jsonStr = "{\"Op\":\"" + Op + "\",\"Key\":\"" + Key + "\",\"Val\":\"" + Val + "\", \"Svc\":\"" + Svc + "\"}";

        return jsonStr;
    }

    /**
     * Build out a full policy statement to test against configuration data
     *
     * @param ontologizer
     * @param dataArray
     * @return
     */
    public static String buildFullJSONPolicyRequestStr(String ontologizer, List dataArray) {
        String jsonStr = "{\"ontology\":\"policy\",\"ontologizer\":\"" + ontologizer + "\",\"data\":[[";

        for (int i = 0; i < dataArray.size(); i++) {
            jsonStr = jsonStr.concat((String) dataArray.get(i));
            if (i != dataArray.size() - 1) {
                jsonStr = jsonStr.concat(",");
            }
        }

        jsonStr += "]]}";

        return jsonStr;
    }

    /**
     * Get a single data item
     *
     * @param cls
     * @param axioms
     * @return
     */
    public static String buildJSONPolicyRequestStr(String cls, String[] axioms) {
        String jsonStr = "{\"Class\":\"" + cls + "\",\"Axioms\":[";
        for (int i = 0; i < axioms.length; i++) {
            if (i != 0) {
                jsonStr += ",";
            }
            jsonStr = jsonStr.concat(axioms[i]);
        }
        jsonStr += "]}";
        return jsonStr;
    }
}
