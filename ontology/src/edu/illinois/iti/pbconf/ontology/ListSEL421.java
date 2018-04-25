package edu.illinois.iti.pbconf.ontology;

import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421;
import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421List;
import java.io.FileInputStream;

class ListSEL421 {

    static void print(SEL421List sel421List) {
        for (SEL421 sel421 : sel421List.getSel421List()) {
            System.out.println("Device name: " + sel421.getName());
            System.out.println("lvlC pwd: " + sel421.getLvlCPwd());
            System.out.println("lvl1B pwd: " + sel421.getLvl1BPwd());
            System.out.println("lvl1A pwd: " + sel421.getLvl1APwd());
            System.out.println("lvl1O pwd: " + sel421.getLvl1OPwd());
            System.out.println("lvl2 pwd: " + sel421.getLvl2Pwd());
            System.out.println("lvl1 pwd: " + sel421.getLvl1Pwd());
            System.out.println("lvl1P [wd: " + sel421.getLvl1PPwd());
            System.out.println("");
        }
    }

    public static void main(String[] args) throws Exception {
        if (args.length != 1) {
            System.err.println("Usage:  ListPeople ADDRESS_BOOK_FILE");
            System.exit(-1);
        }

        SEL421List sel421List = SEL421List.parseFrom(new FileInputStream(args[0]));
        print(sel421List);
    }
}
