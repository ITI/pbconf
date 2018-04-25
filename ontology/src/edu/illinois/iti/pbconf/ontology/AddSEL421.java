package edu.illinois.iti.pbconf.ontology;

import edu.illinois.iti.pbconf.ontology.SEL421Proto.IPAddress;
import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421;
import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421List;
import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421Port5;
import java.io.BufferedReader;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintStream;

class AddSEL421 {

    static IPAddress createIPAddr(String str) {
        IPAddress.Builder ipaddr = IPAddress.newBuilder();
        String[] bytes = str.split("\\.");
        System.out.println(bytes[0] + " " + bytes[1] + " " + bytes[2] + " " + bytes[3]);
        ipaddr.setByte1(Integer.parseInt(bytes[0]));
        ipaddr.setByte2(Integer.parseInt(bytes[1]));
        ipaddr.setByte3(Integer.parseInt(bytes[2]));
        ipaddr.setByte4(Integer.parseInt(bytes[3]));
        return ipaddr.build();
    }

    static SEL421Port5 createSEL421Port5(IPAddress ipaddr) {
        SEL421Port5.Builder sel421Port5 = SEL421Port5.newBuilder();
        sel421Port5.setIpaddr(ipaddr);
        return sel421Port5.build();
    }

    static SEL421 createSEL421(BufferedReader stdin, PrintStream stdout) throws IOException {
        SEL421.Builder sel421 = SEL421.newBuilder();

        stdout.print("Enter device name: ");
        sel421.setName(stdin.readLine());

        stdout.print("Enter lvlC pwd: ");
        sel421.setLvlCPwd(stdin.readLine());

        stdout.print("Enter lvl1B pwd: ");
        sel421.setLvl1BPwd(stdin.readLine());

        stdout.print("Enter lvl1A pwd: ");
        sel421.setLvl1APwd(stdin.readLine());

        stdout.print("Enter lvl1O pwd: ");
        sel421.setLvl1OPwd(stdin.readLine());

        stdout.print("Enter lvl2 pwd: ");
        sel421.setLvl2Pwd(stdin.readLine());

        stdout.print("Enter lvl1 pwd: ");
        sel421.setLvl1Pwd(stdin.readLine());

        stdout.print("Enter lvl1P pwd: ");
        sel421.setLvl1PPwd(stdin.readLine());

        // stdout.print("Enter alarmStt: ");
        // sel421.setAlarmStt(stdin.readLine().equals("true"));
        // stdout.print("Enter accessLoggingStt: ");
        // sel421.setAccessLoggingStt(stdin.readLine().equals("true"));
        // stdout.print("Enter gpsStt: ");
        // sel421.setGpsStt(stdin.readLine().equals("true"));
        // stdout.print("Enter IPAddr: ");
        // SEL421Port5 sel421Port5 = createSEL421Port5(createIPAddr(stdin.readLine()));
        // sel421.setSel421Port5(sel421Port5);
        return sel421.build();
    }

    public static void main(String[] args) throws Exception {
        if (args.length != 1) {
            System.err.println("Usage:  AddSEL421 SEL421_LIST_FILE");
            System.exit(-1);
        }

        SEL421List.Builder sel421List = SEL421List.newBuilder();

        // Read the existing address book.
        try {
            FileInputStream input = new FileInputStream(args[0]);
            try {
                sel421List.mergeFrom(input);
            } finally {
                try {
                    input.close();
                } catch (Throwable ignore) {
                }
            }
        } catch (FileNotFoundException e) {
            System.out.println(args[0] + ": File not found.  Creating a new file.");
        }

        sel421List.addSel421(createSEL421(new BufferedReader(new InputStreamReader(System.in,"UTF-8")), System.out));

        FileOutputStream output = new FileOutputStream(args[0]);
        try {
            sel421List.build().writeTo(output);
        } finally {
            output.close();
        }
    }
}
