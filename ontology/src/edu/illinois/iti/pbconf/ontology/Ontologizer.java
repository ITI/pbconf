/**
 * The Ontologizer class is used to abstract out specific device configuration instances.
 */
package edu.illinois.iti.pbconf.ontology;

import java.util.Set;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassAssertionAxiom;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLOntology;

/**
 * Instances of this abstract class will interpret device agnostic configuration
 * data into device specific ontology axioms for ontological validation.
 *
 * @author Anderson, JD
 */
abstract public class Ontologizer {

    /**
     * Set a property on a device individual
     * @param partialOnt
     * @param device
     * @param Op
     * @param Key
     * @param Val
     * @param Svc
     * @return
     */
    abstract public Boolean setProperty(OWLOntology partialOnt, String device, String Op, String Key, String Val, String Svc);

    /**
     * Determine if an individual has a specific type
     * @param individual
     * @param className
     * @return
     */
    abstract public OWLClassAssertionAxiom getIsA(Individual individual, String className);

    /**
     * Determine if a statement isn't acceptable based on unique information regarding a device
     * An example would be - SEL421 devices don't understand type Person, because they don't have
     * the concept of users.
     * @param Op
     * @param Key
     * @param Val
     * @param Svc
     * @return
     */
    abstract public Boolean isDeviceSpecificValid(String Op, String Key, String Val, String Svc);

    /**
     * Determine what device specific Ontology configuration we'll use This is
     * case insensitive for now
     *
     * @param devType
     * @return
     */
    static Ontologizer getOntologizer(String devType) {
        switch (devType.toUpperCase()) {
            case "SEL421":
                return new SEL421Ontologizer();
            case "LINUX":
                return new LinuxOntologizer();
            default:
                return null;
        }
    }
    
    /**
     * Validates that we support the device
     * @param devType
     * @return 
     */
    static boolean isValidOntologizer(String devType) {
        switch (devType.toUpperCase()) {
            case "SEL421":
                return true;
            case "LINUX":
                return true;
            default:
                return false;
        }
    }
    
    /**
     * Determine if a key is an individual with range Status
     * @param key
     * @return 
     */
    public boolean isStatusKey(String key) {
        //Need to find out if the object property has a range of 'Status' which is {on,off}
        return Ontology.instance().objectPropertyHasRange(key, "Status");
    }
    
    /**
     * Determine if a value has a range of Status {on,off}
     * @param val
     * @return 
     */
    public boolean outsideStatusBounds(String val) {
        return !val.toUpperCase().equals("ON") && !val.toUpperCase().equals("OFF");
    }

    /**
     * Set service_option configuration
     * This is a particularly odd service configuration method and is intentionally limited
     * We can expand this if we find it necessary.
     *
     * @param ont
     * @param device
     * @param Svc
     * @param Key
     * @param Val
     * @return
     * 
     * Key : anonymousftp
     * Val : on | off
     * Svc : 
     * anonymouftp keyword will enable or disable access to a device (unset == off) 
     * 
     * Key : authusers
     * Val : Array of names of users in the ontology, or (on | off)
     * Svc : telnet | ftp | dnp
     * authusers keyword will authorize users for specific services on a target device
     * Right now we only support telnet, ftp and dnp, since those were services specifically mentioned for this
     * 
     * If Val is an array of users, then we will translate (telnet | ftp | dnp) to (authTelnet, authFTP, authDNP)
     * If Val is (on | off), then we will translate (telnet | ftp | dnp) to (hasTelnetStt, hasFTPStt, hasDNP3Stt)
     * which are the actual properties names that were setup in the ontology

     * Assuming it's a list of users : 
     * For each user in the Val property
     * If the individual doesn't already exist, we'll create it and provide it type Person
     * If the individual does exist, but isn't type Person, it will be given type Person
     * 
     * This will result in an individual device having object properties such as
     * authTelnet Joe
     * authTelnet Andy
     * authFTP Andy
     * authDNP Adrian
     * 
     * Basically saying a service is allowed per user
     *
     * From a policy standpoint, we can then set rules like 
     * SEL421 disjointWith ((hasDNP3Stt $value off) $and (authDNP $some Person))
     * Which would say if a device has hasDNP3Stt service set to off, then authorizing 
     * any Person over DNP3 protocol isn't allowed. This will cause an inconsistent ontology and be rejected
     *
     * A standard set of policy for SEL421 devices will probably look something like 
     * rule : An SEL421 device doesn't allow users access over these protocols
     * which translates to SEL421 disjointWith:
     * (authDNP $some Person) $or (authFTP $some Person) $or (authTelnet $some Person)
     */
    public boolean setConfigurationServiceOption(OWLOntology ont, String device, String Svc, String Key, String Val) {
        Individual valInd = null;
        boolean serviceStateSetting = false;
        String[] vals = null;
        switch (Val) {
            case "on":
                valInd = Ontology.instance().getIndividual("on", ont);
                serviceStateSetting = true;
                break;
            case "off":
                valInd = Ontology.instance().getIndividual("off", ont);
                serviceStateSetting = true;
                break;
            default:
                //this command accepts arrays of values, so we'll split on the space delimiter
                vals = Val.split(" ");
                serviceStateSetting = false;
                break;
        }

        if (serviceStateSetting == false) {
            if (Key.toLowerCase().equals("authusers")) {
                boolean validState = true;
                for (String v : vals) {
                    if (validState == false) {
                        continue;
                    }

                    //v = Joe, key is authusers, Svc is the target service
                    String svcProp = "";
                    switch (Svc.toLowerCase()) {
                        case "telnet":
                            svcProp = "authTelnet";
                            break;
                        case "ftp":
                            svcProp = "authFTP";
                            break;
                        case "dnp":
                        case "dnp3":
                            svcProp = "authDNP";
                            break;
                        default:
                            svcProp = "";
                            break;
                    }
                    if ("".equals(svcProp)) {
                        //Unknown property
                        validState = false;
                    } else {
                        Individual individual = Ontology.instance().getIndividual(device, ont);
                        valInd = Ontology.instance().getIndividual(v, ont);
                        IRI propertyIRI = Ontology.instance().getIRI(svcProp);
                        individual.clearProperty(propertyIRI);
                        individual.setProperty(propertyIRI, valInd);
                        validState = true;
                    }
                }
                return validState;
            } else {
                //We can't process this request, only authusers allows something other than on | off
                return false;
            }
        } else {
            //This is a service state setting, meaning we're setting (on | off)
            String key = "";
            if (Key.toLowerCase().equals("authusers")) {
                switch (Svc.toLowerCase()) {
                    case "telnet":
                        key = "hasTelnetStt";
                        break;
                    case "ftp":
                        key = "hasFTPStt";
                        break;
                    case "dnp":
                    case "dnp3":
                        key = "hasDNP3Stt";
                        break;
                    default:
                        key = "";
                        break;
                }
            } else if (Key.toLowerCase().equals("anonymousftp")) {
                key = "hasFTPAnonStt";
            }
            //If we didn't have valid key to work with, reject request
            if ("".equals(key)) {
                return false;
            }
            
            //We have a valid property to set on or off, so set it
            //key can be hasTelnetStt, hasFTPStt, hasFTPAnonStt, or hasDNP3Stt
            Individual individual = Ontology.instance().getIndividual(device, ont);
            IRI propertyIRI = Ontology.instance().getIRI(Key);
            individual.clearProperty(propertyIRI);
            individual.setProperty(propertyIRI, valInd);
            return true;
        }
    }

    /**
     * Set service configuration Service configurations may be either object or
     * data properties depending on what key is passed. If the key matches an
     * existing data property, that is used. If not, it will look for a data
     * property. If that can't be found, it sets a data property. Object
     * properties are limited to what is already in the ontology
     *
     * @param ont
     * @param device
     * @param Key
     * @param Val
     * @return
     */
    public boolean setConfigurationService(OWLOntology ont, String device, String Key, String Val) {
        Key = PBConfParser.translateProperty(Key);
        
        Individual valInd = null;
        switch (Val) {
            case "on":
                valInd = Ontology.instance().getIndividual("on", ont);
                break;
            case "off":
                valInd = Ontology.instance().getIndividual("off", ont);
                break;
            default:
                //This command doesn't accept an array of values, so we only take the first one if there are more than one
                if (Val.contains(" ")) {
                    Val = Val.split(" ")[0];
                }
                valInd = Ontology.instance().getIndividual(Val, ont);
                break;
        }
        Individual individual = Ontology.instance().getIndividual(device, ont);
        IRI propertyIRI = Ontology.instance().getIRI(Key);
        individual.clearProperty(propertyIRI);
        individual.setProperty(propertyIRI, valInd);
        return true;
    }

    /**
     * Set variable configuration Variables are data literals, not object
     * properties They will be set as an XSD:string value
     *
     * @param ont
     * @param device
     * @param Key
     * @param Val
     * @return
     */
    public boolean setConfigurationVariable(OWLOntology ont, String device, String Key, String Val) {
        //Get the individual in the target ontology
        Individual individual = Ontology.instance().getIndividual(device, ont);

        Key = PBConfParser.translateProperty(Key);
        
        //Get the property's IRI and clear existing instances of it
        IRI propertyIRI = Ontology.instance().getIRI(Key);       
        individual.clearProperty(propertyIRI);

        String keyType = PBConfParser.getPropertyRange(Key);
        
        //This is an unknown property, so we're going to have to determine what it is
        //based on the type of value being set.  Since it's configuration 'variable', 
        //it should be a data property
        if ("".equals(keyType)) {
            if (isInteger(Val)) {
                keyType = "integer";
            } else if ("true".equals(Val.toLowerCase()) || "false".equals(Val.toLowerCase())) {
                keyType = "boolean";
            } else if ("".equals(Val)) {
                keyType = "";
            } else {
                keyType = "string";
            }
        }
        
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral valLiteral = null;

        switch (keyType) {
            case "string":
                valLiteral = dataFactory.getOWLLiteral((String) Val);
                break;
            case "integer":
                valLiteral = dataFactory.getOWLLiteral((int) Integer.parseInt(Val));
                break;
            case "boolean":
                valLiteral = dataFactory.getOWLLiteral((Boolean) Boolean.parseBoolean(Val));
                break;
            default:
                //Can't process this type for a data literal
                return false;
        }

        //Set the property to the string value specified
        individual.setProperty(propertyIRI, valLiteral);

        //We made it through the process (not sure if consistent yet at this point), return true
        return true;
    }
    
    /**
     * Helper function to determine if a passed value is a valid integer
     * @param s
     * @return 
     */
    public static boolean isInteger(String s) {
        return isInteger(s, 10);
    }

    /**
     * Helper function to determine if a passed value / radix is a valid integer
     * @param s
     * @param radix
     * @return 
     */
    public static boolean isInteger(String s, int radix) {
        if(s.isEmpty()) return false;
        for(int i = 0; i < s.length(); i++) {
            if(i == 0 && s.charAt(i) == '-') {
                if(s.length() == 1) return false;
                else continue;
            }
            if(Character.digit(s.charAt(i),radix) < 0) return false;
        }
        return true;
    }

    /**
     * Set password configuration Passwords are data literals, not object
     * properties Passwords will always be set as an XSD:string value If the
     * selected individual doesn't exist, it will be created
     * 
     * @param ont
     * @param device
     * @param Key
     * @param Val
     * @return
     */
    public boolean setConfigurationPassword(OWLOntology ont, String device, String Key, String Val) {
        //Get the individual in the target ontology
        Individual individual = Ontology.instance().getIndividual(device, ont);

        Key = PBConfParser.translateProperty(Key);
        //Get the property's IRI and clear existing instances of it
        IRI propertyIRI = Ontology.instance().getIRI(Key);            
        individual.clearProperty(propertyIRI);

        //Set the property to the string value specified
        individual.setProperty(propertyIRI, Val);

        //We made it through the process (not sure if consistent yet at this point), return true
        return true;
    }
    
    /**
     * Set a devices class
     * @param ont
     * @param device
     * @param Val
     * @return 
     */
    public boolean setConfigurationType(OWLOntology ont, String device, String Val) {
        try {
            //Get the individual in the target ontology
            Individual individual = Ontology.instance().getIndividual(device, ont);

            //Set the devices class as needed (will likely be SEL421)
            individual.setClass(Ontology.instance().getIRI(Val));
 
            //We made it through the process (not sure if consistent yet at this point), return true
            return true;
        } catch (Exception ex) {
            //We failed to set the property somewhere in the process, return false
            return false;
        }
    }

    /**
     * LinuxOntologizer handles configuration for a standard Linux box.
     * Currently, there are no device specific restrictions known.
     */
    private static class LinuxOntologizer extends Ontologizer {

        /**
         * Add a property configuration
         *
         * @param Op
         * @param Key
         * @param Val
         * @param Svc
         * @return
         */
        @Override
        public Boolean setProperty(OWLOntology partialOnt, String device, String Op, String Key, String Val, String Svc) {
            //Start by making sure we don't violate rules set specific to this ontologizer
            boolean specificValid = this.isDeviceSpecificValid(Op, Key, Val, Svc);
            
            if (specificValid == false) {
                return false;
            }
            
            //We'll always be working in the configuration ontology, so select that
            OntologyConfig cfg = Ontology.instance().getConfig();
            String prefix = cfg.getConfigPrefix();
            boolean state = false;

            switch (Op.toLowerCase()) {
                case "service_option":
                    state = setConfigurationServiceOption(partialOnt, device, Svc, Key, Val);
                    break;
                case "service":
                    state = setConfigurationService(partialOnt, device, Key, Val);
                    break;
                case "variable":
                    //Need to also allow type set from here as a key word (Op:variable,Key:type,Val:${type}
                    if (Key.toLowerCase().equals("type")) {
                        state = setConfigurationType(partialOnt, device, Val);
                    } else {
                        state = setConfigurationVariable(partialOnt, device, Key, Val);
                    }
                    break;
                case "password":
                    //If we're setting a password, we know we will be setting 
                    //a string literal value, so we can treat the val as such
                    //resulting an a {device} has {leveled password} {string literal}
                    state = setConfigurationPassword(partialOnt, device, Key, Val);
                    break;
                case "type":
                    state = setConfigurationType(partialOnt, device, Val);
                    break;
                default:
                    //We don't currently support any additional operations
                    state = false;
                    break;
            }

            return state;
        }

        /**
         * Determine if an individual is of a class
         *
         * @param individual
         * @param className
         * @return
         */
        @Override
        public OWLClassAssertionAxiom getIsA(Individual individual, String className) {
            IRI classIRI = Ontology.instance().getIRI(className);
            OWLClass owlClass = Ontology.instance().getDataFactory().getOWLClass(classIRI);
            OWLIndividual owlInd = individual.getOWLIndividual();
            OWLClassAssertionAxiom axiom = Ontology.instance().getDataFactory().getOWLClassAssertionAxiom(owlClass, owlInd);
            return axiom;
        }

        /**
         * Determine if a request breaks any device specific rules
         *
         * @param Op
         * @param Key
         * @param Val
         * @param Svc
         * @return
         */
        @Override
        public Boolean isDeviceSpecificValid(String Op, String Key, String Val, String Svc) {
            /*
             * This will be extended if any Linux specific restrictions become apparent
             */
            return true;
        }
    }

    /*
     This private impl cannot be directly constructed. 
     Must use the factory method. 
     */
    private static class SEL421Ontologizer extends Ontologizer {

        /**
         * We receive an operation type, a Key we're operating on, and a Val
         * which may consist of "on", "off", or one or more space delimited
         * items we'll need to set
         *
         * Op can be service, variable, password, or service_option Svc will
         * only exist as an extra value when service_option is used
         *
         * @param Op
         * @param Key
         * @param Val
         * @param Svc
         * @return
         *
         */
        @Override
        public Boolean setProperty(OWLOntology partialOnt, String device, String Op, String Key, String Val, String Svc) {
            //Start by making sure we don't violate rules set specific to this ontologizer
            boolean specificValid = this.isDeviceSpecificValid(Op, Key, Val, Svc);
            
            if (specificValid == false) {
                return false;
            }
            
            //We'll always be working in the configuration ontology, so select that
            OntologyConfig cfg = Ontology.instance().getConfig();
            String prefix = cfg.getConfigPrefix();
            boolean state = false;

            switch (Op.toLowerCase()) {
                case "service_option":
                    state = setConfigurationServiceOption(partialOnt, device, Svc, Key, Val);
                    break;
                case "service":
                    state = setConfigurationService(partialOnt, device, Key, Val);
                    break;
                case "variable":
                    //Need to also allow type set from here as a key word (Op:variable,Key:type,Val:${type}
                    if (Key.toLowerCase().equals("type")) {
                        state = setConfigurationType(partialOnt, device, Val);
                    } else {
                        state = setConfigurationVariable(partialOnt, device, Key, Val);
                    }
                    break;
                case "password":
                    //If we're setting a password, we know we will be setting 
                    //a string literal value, so we can treat the val as such
                    //resulting an a {device} has {leveled password} {string literal}
                    state = setConfigurationPassword(partialOnt, device, Key, Val);
                    break;
                case "type":
                    state = setConfigurationType(partialOnt, device, Val);
                    break;
                default:
                    //We don't currently support any additional operations
                    state = false;
                    break;
            }

            return state;
        }

        @Override
        public OWLClassAssertionAxiom getIsA(Individual individual, String className) {
            IRI classIRI = Ontology.instance().getIRI(className);
            OWLClass owlClass = Ontology.instance().getDataFactory().getOWLClass(classIRI);
            OWLIndividual owlInd = individual.getOWLIndividual();
            OWLClassAssertionAxiom axiom = Ontology.instance().getDataFactory().getOWLClassAssertionAxiom(owlClass, owlInd);
            return axiom;
        }

        /**
         * Determine if a request breaks any device specific rules
         *
         * @param Op
         * @param Key
         * @param Val
         * @param Svc
         * @return
         */
        @Override
        public Boolean isDeviceSpecificValid(String Op, String Key, String Val, String Svc) {
            /**
             * Current known restrictions : 1.) Any user specific configuration
             * is considered invalid on an SEL421 device, so automatically
             * reject This occurs when individuals specified in Val have type
             * "Person", or type "Role" and the Op is a service_option
             */
            OntologyConfig cfg = Ontology.instance().getConfig();
            String prefix = cfg.getConfigPrefix();
            OWLOntology ont = Ontology.instance().getOntology(prefix);
            if (null != Op) switch (Op) {
                case "service_option":
                    String[] vals = Val.split(" ");
                    boolean requestContainsUser = false;
                    for (String val : vals) {
                        Individual ind = Ontology.instance().getIndividual(val, ont);
                        OWLIndividual oInd = ind.getOWLIndividual();
                        Set<OWLClassExpression> expressions = oInd.getTypes(ont);
                        
                        for (OWLClassExpression expression : expressions) {
                            String type = expression.toString();
                            type = type.split("#")[1].replace(">", "");
                            if ("Person".equals(type)) {
                                requestContainsUser = true;
                            }
                        }
                    }   if (requestContainsUser == true) {
                    return false;
                    }   break;
                case "service":
                    if (isStatusKey(Key) && outsideStatusBounds(Val)) {
                    return false;
                }   break;
            }
            
            return true;
        }
    }
}
